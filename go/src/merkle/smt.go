package merkle

import (
	"encoding/base64"
	"fmt"
	"encoding/base64"
	"crypto/sha256"
	"bytes"
	"sync"
	"sort"
	"strconv"
	"runtime"
)

const TREE_DEPTH int = 256

const BATCH_READ_SIZE int = 50

const BATCH_PERCOLATE_SIZE int = 50

type digest = []byte

type SparseMerkleTree struct {
	depth int
	cache map[string]*digest
	rootDigest digest
	empty_cache map[int]digest
	conflicts map[string]*SyncBool
}

type SyncBool struct {
	lock *sync.Mutex
	visited bool
}

func (T *SparseMerkleTree) GetRoot() (digest) {
	return T.rootDigest
}

func (T *SparseMerkleTree) getEmpty(n int) (digest) {
	if (len(T.empty_cache) <= n) {
		t := T.getEmpty(n - 1)
		T.empty_cache[n] = hashDigest(append(t[:], t[:]...))
	} 
	return T.empty_cache[n]
}

func hashDigest(Data []byte) (digest) {
	H := sha256.New()
	H.Write(Data)
	return H.Sum(nil)
}

func getParent(nodeID string) (string) {
	length := len(nodeID)
	if (length == 0) {
		return nodeID
	}

	return nodeID[:length - 1]
}

func getSibling(nodeID string) (string, bool) {
	length := len(nodeID)
	if (length == 0) {
		return nodeID, false
	}

	
	lastBit := nodeID[length - 1]
	siblingID := nodeID[:length - 1]

	var isLeft bool

	if (lastBit == byte('0')) {
		siblingID += "1"
		isLeft = false
	} else {
		siblingID += "0"
		isLeft = true
	}

	return siblingID, isLeft
}

func (T *SparseMerkleTree) getEmptyAncestor(nodeID string) (string) {
	currID := nodeID
	prevID := currID
	for (len(currID) > 0) {
		if _, ok := T.cache[currID]; ok {
			break
		}
		prevID = currID
		currID = getParent(currID)
	}
	return prevID
}

/* Assume index is valid (for now).

FIXME: What are the error cases where we return error/False?
*/
func (T *SparseMerkleTree) Insert(index string, data string) (bool) {
	dig, _ := base64.StdEncoding.DecodeString(data)
	hash := hashDigest(dig)
	T.cache[index] = &hash

	//FIXME: Actual copy?
	var currID string
	for currID = index; len(currID) > 0; {
		// Get both the parent and sibling IDs
		siblingID, isLeft := getSibling(currID)
		parentID := getParent(currID)

		// Get the digest of the current node and sibling
		currDigestPointer := T.cache[currID] // currID will always be in cache
		siblingDigestPointer, ok := T.cache[siblingID]
		var siblingDigest digest
		if !ok {
			siblingDigest = T.getEmpty(len(siblingID))
		} else {
			siblingDigest = *siblingDigestPointer
		}
		currDigest := *currDigestPointer

		// Hash the digests of the left and right children
		var parentDigest digest
		if isLeft {
			parentDigest = hashDigest(append(siblingDigest, currDigest...))
		} else {
			parentDigest = hashDigest(append(currDigest, siblingDigest...))
		}
		T.cache[parentID] = &parentDigest

		// Traverse up the tree by making the current node the parent node
		currID = parentID
	}
	rootDigestPointer := T.cache[currID]
	T.rootDigest = *rootDigestPointer
	return true
}

func (T *SparseMerkleTree) preloadCopaths(transactions BatchedTransaction, read chan []*CoPathPair, db *angelaDB) (bool, error) {
	// compute copaths that need to be pulled in (need ids alone)
	copaths := make(map[string]bool)
	transactionLength := len(transactions)
	var currID string

	for i := 0; i < transactionLength; i++ {
		currID = transactions[i].Id
		// Our stopping condition is length > 0 so we don't add the root to the copath
		for ; len(currID) > 0; currID = getParent(currID) {
			siblingID, _ := getSibling(currID)
			copaths[siblingID] = true
		}
	}

	ids := make([]string, 0, len(copaths))
    for k := range copaths {
        ids = append(ids, k)
    }

	copathPairs, err := db.retrieveLatestCopathDigests(ids)
	if err != nil {
		fmt.Println(err)
		read <- copathPairs
		return false, err
	}
	read <- copathPairs
	return true, nil
}

func auroraWriteback(ch chan []*CoPathPair, quit chan bool, db *angelaDB) {
	var epochNumber int64 = 1
    for {
    	select {
    	case delta := <-ch:
    		// write back to aurora of the changelist from the channel
    		// fmt.Println("Changelist")
    		// fmt.Println(delta)
    		numRowsAffected, err := db.insertChangeList(delta, epochNumber)
    		if err != nil {
    			fmt.Println(err)
    		}
    		fmt.Println("Printing the number of rows affected")
    		fmt.Println(numRowsAffected)
    	case <-quit:
    		fmt.Println("Done Writing")
    		return
    	}
    }
}

func (T *SparseMerkleTree) BatchInsert(transactions BatchedTransaction) (bool, error) {
	readChannel := make(chan []*CoPathPair)

	readDB, err := GetReadAngelaDB()
	if err != nil {
		panic(err)
	}
	defer readDB.Close()
	writeDB, err := GetWriteAngelaDB()
	if err != nil {
		panic(err)
	}
	defer writeDB.Close()


	// fmt.Println("Cache before preload")
	// fmt.Println(T.cache)
	sort.Sort(BatchedTransaction(transactions))

	for i := 0; i < len(transactions); i+=BATCH_READ_SIZE {
		go T.preloadCopaths(transactions[i:min(i+BATCH_READ_SIZE, len(transactions))], readChannel, readDB)
	}

	for i := 0; i < len(transactions); i+=BATCH_READ_SIZE {
		copathPairs := <-readChannel
		for j := 0; j < len(copathPairs); j++ {
			T.cache[copathPairs[j].ID] = &copathPairs[j].Digest
		}
	}
	// fmt.Println("Cache after preload")
	// fmt.Println(T.cache)
	T.conflicts, err = findConflicts(transactions)
	if err != nil {
		return false, err
	}

	for _, Transaction := range transactions {
		for currID := Transaction.Id; currID != ""; currID = getParent(currID) {
			placeHolder := T.getEmpty(0)
			T.cache[currID] = &placeHolder
		}
	}
	placeHolder := T.getEmpty(0)
	T.cache[""] = &placeHolder

	var wg sync.WaitGroup
	ch := make(chan []*CoPathPair)
	quit := make(chan bool)

	go auroraWriteback(ch, quit, writeDB)

	for i:=0; i<len(transactions); i++ {
		wg.Add(1)
		go T.percolate(transactions[i].Id, transactions[i].Data, &wg, ch)
	}
	wg.Wait()
	quit <- true
	rootDigestPointer := T.cache[""]
	T.rootDigest = *rootDigestPointer

	return true, nil
}

func (T *SparseMerkleTree) batch2Insert(transactions BatchedTransaction) (bool, error) {
	readChannel := make(chan []*CoPathPair)
	readDB, err := GetReadAngelaDB()
	if err != nil {
		panic(err)
	}
	defer readDB.Close()
	writeDB, err := GetWriteAngelaDB()
	if err != nil {
		panic(err)
	}
	sort.Sort(BatchedTransaction(transactions))
	for i := 0; i < len(transactions); i+=BATCH_READ_SIZE {
		go T.preloadCopaths(transactions[i:min(i+BATCH_READ_SIZE, len(transactions))], readChannel, readDB)
	}

	for i := 0; i < len(transactions); i+=BATCH_READ_SIZE {
		copathPairs := <-readChannel
		for j := 0; j < len(copathPairs); j++ {
			T.cache[copathPairs[j].ID] = &copathPairs[j].Digest
		}
	}

	T.conflicts, err = findConflicts(transactions)
	if err != nil {
		return false, err
	}

	for _, Transaction := range transactions {
		for currID := Transaction.Id; currID != ""; currID = getParent(currID) {
			placeHolder := T.getEmpty(0)
			T.cache[currID] = &placeHolder
		}
	}
	placeHolder := T.getEmpty(0)
	T.cache[""] = &placeHolder

	var wg sync.WaitGroup
	lenTrans := len(transactions)
	ch := make(chan []*CoPathPair)
	quit := make(chan bool)

	go auroraWriteback(ch, quit, writeDB)

	stepSize := len(transactions) / runtime.GOMAXPROCS(0)
	for i:=0; i<len(transactions); i+=stepSize {
		wg.Add(1)
		go T.batchPercolate(transactions[i:min(i+stepSize, lenTrans)], &wg, ch)
	}
	wg.Wait()

	quit <- true
	rootDigestPointer := T.cache[""]
	T.rootDigest = *rootDigestPointer

	return true, nil
}

func (T *SparseMerkleTree) percolate(index string, Data string, wg *sync.WaitGroup, ch chan []*CoPathPair) (bool, error) {
	defer wg.Done()

	changeList := make([]*CoPathPair, 0)

	dig, _ := base64.StdEncoding.DecodeString(data)
	changeList = append(changeList, &CoPathPair{ID: index, Digest: hashDigest(dig)})
	
	//TODO: You should not hash the value passed in if it not a leaf ie in the root tree
	hash := hashDigest(dig)
	indexPointer := T.cache[index]
	*indexPointer = hash

	var currID string
	for currID = index; len(currID) > 0; {
		// Get both the parent and sibling IDs
		siblingID, isLeft := getSibling(currID)
		parentID := getParent(currID)

		//conflict check
		if T.isConflict(parentID) {
			ch <- changeList
			return true, nil
		}

		// Get the digest of the current node and sibling
		currDigestPointer := T.cache[currID] // currID will always be in cache
		siblingDigestPointer, ok := T.cache[siblingID]
		var siblingDigest digest
		if !ok {
			siblingDigest = T.getEmpty(len(siblingID))
		} else {
			siblingDigest = *siblingDigestPointer
		}
		currDigest := *currDigestPointer

		// Hash the digests of the left and right children
		var parentDigest digest
		if isLeft {
			parentDigest = hashDigest(append(siblingDigest, currDigest...))
		} else {
			parentDigest = hashDigest(append(currDigest, siblingDigest...))
		}
		changeList = append(changeList, &CoPathPair{ID: parentID, Digest: parentDigest})
		parentDigestPointer := T.cache[parentID]
		*parentDigestPointer = parentDigest

		// Traverse up the tree by making the current node the parent node
		currID = parentID
	}
	rootDigestPointer := T.cache[currID]
	T.rootDigest = *rootDigestPointer
	ch <- changeList
	return true, nil
}

func (T *SparseMerkleTree) batchPercolate(transactions BatchedTransaction, wg *sync.WaitGroup, ch chan []*CoPathPair) (bool, error) {
	defer wg.Done()
	changeList := make([]*CoPathPair, 0)
	
	for _, trans := range transactions {
		index := trans.Id
		Data := trans.Data

		//TODO: You should not hash the value passed in if it not a leaf ie in the root tree
		dig, _ := base64.StdEncoding.DecodeString(data)
		hash := hashDigest(dig)
		indexPointer := T.cache[index]
		*indexPointer = hash
		changeList = append(changeList, &CoPathPair{ID: index, Digest: hash})

		var currID string
		for currID = index; len(currID) > 0; {
			// Get both the parent and sibling IDs
			siblingID, isLeft := getSibling(currID)
			parentID := getParent(currID)

			//conflict check
			if T.isConflict(parentID) {
				continue
			}

			// Get the digest of the current node and sibling
			currDigestPointer := T.cache[currID] // currID will always be in cache
			siblingDigestPointer, ok := T.cache[siblingID]
			var siblingDigest digest
			if !ok {
				siblingDigest = T.getEmpty(len(siblingID))
			} else {
				siblingDigest = *siblingDigestPointer
			}
			currDigest := *currDigestPointer

			// Hash the digests of the left and right children
			var parentDigest digest
			if isLeft {
				parentDigest = hashDigest(append(siblingDigest, currDigest...))
			} else {
				parentDigest = hashDigest(append(currDigest, siblingDigest...))
			}
			parentDigestPointer := T.cache[parentID]
			*parentDigestPointer = parentDigest
			changeList = append(changeList, &CoPathPair{ID: parentID, Digest: parentDigest})

			// Traverse up the tree by making the current node the parent node
			currID = parentID
		}
		rootDigestPointer := T.cache[currID]
		T.rootDigest = *rootDigestPointer

		continue
	}
	ch <- changeList
	return true, nil
}

func (T *SparseMerkleTree) isConflict(index string) (bool) {
	if val, ok := T.conflicts[index]; ok { 
		val.lock.Lock()
		val := T.conflicts[index]
		defer val.lock.Unlock()

		if !val.visited {
			val.visited = true
			return true
		}
	}

	return false
}

func (T *SparseMerkleTree) CGenerateProof(index string) ([]string) {
	proof := T.generateProof(index)
	resultLength := len(proof.coPath)*2+3
	results := make([]string, resultLength)
    results[0] = strconv.FormatBool(bool(proof.proofType))
    results[1] = proof.queryID
    results[2] = proof.proofID
    for i := 3; i < resultLength; i += 2 {
    	results[i] = proof.coPath[i-3].ID
    	results[i+1] = base64.StdEncoding.EncodeToString(proof.coPath[i-3].digest)
    }
    return results
}

func (T *SparseMerkleTree) GenerateProof(index string) (Proof) {
	proofResult := Proof{}
	proofResult.QueryID = index

	var proof_t ProofType
	var currID string
	_, ok := T.cache[index]
	if !ok {
		proof_t = NONMEMBERSHIP
		currID = T.getEmptyAncestor(index)
	} else {
		proof_t = MEMBERSHIP
		currID = index
	}
	fmt.Println("Call to empty")
	fmt.Println(base64.StdEncoding.EncodeToString(T.getEmpty(TREE_DEPTH - len(currID))))
	proofResult.ProofType = proof_t
	proofResult.ProofID = currID
	CoPath := make([]CoPathPair, 0)

	// Our stopping condition is length > 0 so we don't add the root to the copath
	for ; len(currID) > 0; currID = getParent(currID) {
		// Append the sibling to the copath and advance current node
		siblingID, _ := getSibling(currID)
		siblingDigestPointer, ok := T.cache[siblingID]
		var siblingDigest digest
		if !ok {
			siblingDigest = T.getEmpty(TREE_DEPTH - len(siblingID))
		} else {
			siblingDigest = *siblingDigestPointer
		}

		CoPathNode := CoPathPair{siblingID, siblingDigest}
		CoPath = append(CoPath, CoPathNode)
	}

	proofResult.CoPath = CoPath
	return proofResult
}

func (T *SparseMerkleTree) verifyProof(proof Proof) (bool) {
	// If proof of nonmembership, first make sure that there is a prefix match
	ProofIDLength := len(proof.ProofID)
	if proof.ProofType == NONMEMBERSHIP {
		if ProofIDLength > len(proof.QueryID) {
			return false
		}

		for i := 0; i < ProofIDLength; i++ {
			if proof.ProofID[i] != proof.QueryID[i] {
				return false
			}
		}
	}

	rootDigestPointer, ok := T.cache[proof.ProofID]
	var rootDigest digest
	if !ok {
		rootDigest = T.getEmpty(TREE_DEPTH - ProofIDLength)
	} else {
		rootDigest = *rootDigestPointer
	}

	for i := 0; i < len(proof.CoPath); i++ {
		currNode := proof.CoPath[i]
		if currNode.ID[len(currNode.ID) - 1] == '0' {
			rootDigest = hashDigest(append(currNode.Digest, rootDigest...))
		} else {
			rootDigest = hashDigest(append(rootDigest, currNode.Digest...))
		}
	}
	return bytes.Equal(rootDigest, T.rootDigest)
}

// leaves must be sorted before findConflicts is called
func findConflicts(leaves []*Transaction) (map[string]*SyncBool, error) {
	var conflicts = make(map[string]*SyncBool)
	for i := 1; i < len(leaves); i++ {
		x, y := leaves[i-1].Id, leaves[i].Id
		k := len(x)
		for idx := 0; idx < len(leaves); idx++ {
			var a, b byte = x[idx], y[idx]
			if a != b {
				k = idx
				break
			}
		}
		conflicts[x[0:k]] = &SyncBool{lock: &sync.Mutex{}, visited: false}
	}
	return conflicts, nil
}

func MakeTree() (*SparseMerkleTree) {
	T := SparseMerkleTree{} 
	T.depth = TREE_DEPTH
	T.cache = make(map[string]*digest)
	T.empty_cache = make(map[int]digest)
	dig, _ := base64.StdEncoding.DecodeString("")
	T.empty_cache[0] = hashDigest(dig)  
	T.rootDigest = T.getEmpty(TREE_DEPTH) // FIXME: Should this be the case for an empty tree?
	T.conflicts = make(map[string]*SyncBool)
	return &T
}
