package merkle

import (
	"encoding/base64"
	"fmt"
	"crypto/sha256"
	"bytes"
	"sync"
	_ "sort"
	"strconv"
	"errors"
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
	prefix string
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
			siblingDigest = T.getEmpty(T.depth - len(siblingID))
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
		currID = transactions[i].ID
		// Our stopping condition is length > 0 so we don't add the root to the copath
		for ; len(currID) > 0; currID = getParent(currID) {
			siblingID, _ := getSibling(currID)
			copaths[siblingID] = true
		}
	}

	ids := make([]string, 0, len(copaths))
    for k := range copaths {
        ids = append(ids, T.prefix + k)
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

func auroraWriteback(ch chan []*CoPathPair, quit chan bool, db *angelaDB, prefix string, epochNumber uint64) {
    for {
    	select {
    	case delta := <-ch:
    		// write back to aurora of the changelist from the channel
    		// fmt.Println("Changelist")
    		// fmt.Println(delta)
    		for i:=0; i < len(delta); i++ {
    			delta[i].ID = prefix + delta[i].ID
    		}
    		_, err := db.insertChangeList(delta, epochNumber)
    		if err != nil {
    			fmt.Println(err)
    		}
    		// fmt.Println("Printing the number of rows affected")
    		// fmt.Println(numRowsAffected)
    	case <-quit:
    		fmt.Println("Done Writing")
    		return
    	}
    }
}

func (T *SparseMerkleTree) BatchInsert(transactions BatchedTransaction, epochNumber uint64) (string, error) {
	// readChannel := make(chan []*CoPathPair)
	// readDB, err := GetReadAngelaDB()
	// if err != nil {
	// 	panic(err)
	// }
	// defer readDB.Close()
	// writeDB, err := GetWriteAngelaDB()
	// if err != nil {
	// 	panic(err)
	// }
	// defer writeDB.Close()


	// fmt.Println("Cache before preload")
	// fmt.Println(T.cache)
	// sort.Sort(BatchedTransaction(transactions))

	for _, Transaction := range transactions {
		for currID := Transaction.ID; currID != ""; currID = getParent(currID) {
			placeHolder := T.getEmpty(T.depth)
			T.cache[currID] = &placeHolder
		}
	}
	placeHolder := T.getEmpty(T.depth)
	T.cache[""] = &placeHolder

	// for i := 0; i < len(transactions); i+=BATCH_READ_SIZE {
	// 	go T.preloadCopaths(transactions[i:min(i+BATCH_READ_SIZE, len(transactions))], readChannel, readDB)
	// }

	// for i := 0; i < len(transactions); i+=BATCH_READ_SIZE {
	// 	copathPairs := <-readChannel
	// 	for j := 0; j < len(copathPairs); j++ {
	// 		T.cache[copathPairs[j].ID[len(T.prefix):]] = &copathPairs[j].Digest
	// 	}
	// }
	// fmt.Println("Cache after preload")
	// fmt.Println(T.cache)
	err := errors.New("")
	err = nil
	T.conflicts, err = findConflicts(transactions)
	if err != nil {
		return "", err
	}

	var wg sync.WaitGroup
	ch := make(chan []*CoPathPair)
	//quit := make(chan bool)

	// go auroraWriteback(ch, quit, writeDB, T.prefix, epochNumber)

	for i:=0; i<len(transactions); i++ {
		wg.Add(1)
		go T.percolate(transactions[i].ID, transactions[i].Data, &wg, ch)
	}
	wg.Wait()
	// quit <- true
	rootDigestPointer := T.cache[""]
	T.rootDigest = *rootDigestPointer

	return base64.StdEncoding.EncodeToString(T.rootDigest), nil
}

func (T *SparseMerkleTree) percolate(index string, data string, wg *sync.WaitGroup, ch chan []*CoPathPair) (bool, error) {
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
			//ch <- changeList
			return true, nil
		}

		// Get the digest of the current node and sibling
		currDigestPointer := T.cache[currID] // currID will always be in cache
		siblingDigestPointer, ok := T.cache[siblingID]
		var siblingDigest digest
		if !ok {
			siblingDigest = T.getEmpty(T.depth - len(siblingID))
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
		// fmt.Println(base64.StdEncoding.EncodeToString(parentDigest))
		changeList = append(changeList, &CoPathPair{ID: parentID, Digest: parentDigest})
		parentDigestPointer := T.cache[parentID]
		*parentDigestPointer = parentDigest

		// Traverse up the tree by making the current node the parent node
		currID = parentID
	}
	rootDigestPointer := T.cache[currID]
	T.rootDigest = *rootDigestPointer
	//ch <- changeList
	return true, nil
}

func (T *SparseMerkleTree) batchPercolate(transactions BatchedTransaction, wg *sync.WaitGroup, ch chan []*CoPathPair) (bool, error) {
	defer wg.Done()
	changeList := make([]*CoPathPair, 0)
	
	for _, trans := range transactions {
		index := trans.ID
		data := trans.Data

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
				siblingDigest = T.getEmpty(T.depth - len(siblingID))
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

func (T *SparseMerkleTree) GetLatestRoot() (string) {
	readDB, err := GetReadAngelaDB()
	if err != nil {
		panic(err)
	}
	defer readDB.Close()
	rootId := []string{""}
	copathPairs, err := readDB.retrieveLatestCopathDigests(rootId)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return base64.StdEncoding.EncodeToString(copathPairs[0].Digest)
}

func (T *SparseMerkleTree) CGenerateProof(index string) ([]string) {
	proof := T.GenerateProofDB(index)
	results := make([]string, proof.ProofLength)
    results[0] = strconv.FormatBool(bool(proof.ProofType))
    results[1] = proof.QueryID
    results[2] = proof.ProofID
    results[3] = strconv.FormatInt(int64(proof.ProofLength), 10)
    for i := 4; i < proof.ProofLength; i += 2 {
    	results[i] = proof.CoPath[(i-4)/2].ID
    	results[i+1] = base64.StdEncoding.EncodeToString(proof.CoPath[(i-4)/2].Digest)
    }
    return results
}

func (T *SparseMerkleTree) GenerateProofDB(index string) (Proof) {
	readDB, err := GetReadAngelaDB()
	if err != nil {
		panic(err)
	}
	defer readDB.Close()
	proofResult := Proof{}
	proofResult.QueryID = index
	
	var proof_t ProofType
	var currID string
	var startingIndex string
	indexCheck := []string{index}
	fmt.Println("Index to check", indexCheck)
	indexPair, err := readDB.retrieveLatestCopathDigests(indexCheck)
	ok := err == nil && len(indexPair) > 0

	// _, ok := T.cache[index]
	if !ok {
		proof_t = NONMEMBERSHIP
		ancestorIds := make([]string, 0)
		ancestor := index
		// Our stopping condition is length > 0 so we don't add the root to the copath
		for ; len(ancestor) > 0; ancestor = getParent(ancestor) {
			ancestorIds = append(ancestorIds, ancestor)
		}
	    fmt.Println("number of ancestors", len(ancestorIds))

		copathPairs, err := readDB.retrieveLatestCopathDigests(ancestorIds)
		if err != nil {
			fmt.Println(err)
		}
		for j := 0; j < len(copathPairs); j++ {
			T.cache[copathPairs[j].ID] = &copathPairs[j].Digest
		}
		startingIndex = T.getEmptyAncestor(index)
	} else {
		proof_t = MEMBERSHIP
		startingIndex = index
	}
	fmt.Println("startingIndex is", startingIndex)
	proofResult.ProofType = proof_t
	proofResult.ProofID = startingIndex
	CoPath := make([]CoPathPair, 0)

	currID = startingIndex
	ids := make([]string, 0)
	// Our stopping condition is length > 0 so we don't add the root to the copath
	for ; len(currID) > 0; currID = getParent(currID) {
		siblingID, _ := getSibling(currID)
		ids = append(ids, siblingID)
	}

	copathPairs, err := readDB.retrieveLatestCopathDigests(ids)
	if err != nil {
		fmt.Println(err)
		proofResult.CoPath = CoPath
		return proofResult
	}

	for j := 0; j < len(copathPairs); j++ {
		T.cache[copathPairs[j].ID] = &copathPairs[j].Digest
	}

	currID = startingIndex
	// Our stopping condition is length > 0 so we don't add the root to the copath
	for ; len(currID) > 0; currID = getParent(currID) {
		// Append the sibling to the copath and advance current node
		siblingID, _ := getSibling(currID)
		siblingDigestPointer, ok := T.cache[siblingID]
		var siblingDigest digest
		if !ok {
			siblingDigest = T.getEmpty(T.depth - len(siblingID))
		} else {
			siblingDigest = *siblingDigestPointer
		}
		CoPathNode := CoPathPair{siblingID, siblingDigest}
		CoPath = append(CoPath, CoPathNode)
	}
	fmt.Println("Length of CoPath", len(CoPath))
	// 4 metadata fields and 2 times the copath length
	proofResult.ProofLength = len(CoPath)*2+4
	proofResult.CoPath = CoPath
	return proofResult	
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
			siblingDigest = T.getEmpty(T.depth - len(siblingID))
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
		rootDigest = T.getEmpty(T.depth - ProofIDLength)
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
		x, y := leaves[i-1].ID, leaves[i].ID
		k := len(x)
		// This was originally len(leaves)...
		for idx := 0; idx < len(x); idx++ {
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

func MakeTree(prefix string) (*SparseMerkleTree) {
	T := SparseMerkleTree{} 
	T.depth = TREE_DEPTH - len(prefix)
	T.cache = make(map[string]*digest)
	T.empty_cache = make(map[int]digest)
	dig, _ := base64.StdEncoding.DecodeString("")
	T.empty_cache[0] = hashDigest(dig)  
	T.rootDigest = T.getEmpty(T.depth) // FIXME: Should this be the case for an empty tree?
	T.conflicts = make(map[string]*SyncBool)
	T.prefix = prefix
	return &T
}
