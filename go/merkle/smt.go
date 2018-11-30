package merkle

import (
	"fmt"
	"crypto/sha256"
	"bytes"
	"sync"
	"sort"
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

func (T *SparseMerkleTree) getEmpty(n int) (digest) {
	if (len(T.empty_cache) <= n) {
		t := T.getEmpty(n - 1)
		T.empty_cache[n] = hashDigest(append(t[:], t[:]...))
	} 
	return T.empty_cache[n]
}

func hashDigest(data []byte) (digest) {
	H := sha256.New()
	H.Write(data)
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
func (T *SparseMerkleTree) insert(index string, data string) (bool) {
	hash := hashDigest([]byte(data))
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

func (T *SparseMerkleTree) preloadCopaths(transactions batchedTransaction, read chan []*CoPathPair) (bool, error) {
	// compute copaths that need to be pulled in (need ids alone)
	copaths := make(map[string]bool)
	transactionLength := len(transactions)
	var currID string

	for i := 0; i < transactionLength; i++ {
		currID = transactions[i].id
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

	copathPairs, err := retrieveCopaths(ids)
	if err != nil {
		fmt.Println(err)
		read <- copathPairs
		return false, err
	}
	read <- copathPairs
	return true, nil
}

func auroraWriteback(ch chan []*CoPathPair, quit chan bool) {
	var epochNumber int64 = 1
    for {
    	select {
    	case delta := <-ch:
    		// write back to aurora of the changelist from the channel
    		// fmt.Println("Changelist")
    		// fmt.Println(delta)
    		numRowsAffected, err := writeChangeList(delta, epochNumber)
    		if err != nil {
    			fmt.Println("Errored with %v", err)
    		}
    		fmt.Println("Printing the number of rows affected")
    		fmt.Println(numRowsAffected)
    	case <-quit:
    		fmt.Println("Done Writing")
    		return
    	}
    }
}

func (T *SparseMerkleTree) batchInsert(transactions batchedTransaction) (bool, error) {
	readChannel := make(chan []*CoPathPair)
	// fmt.Println("Cache before preload")
	// fmt.Println(T.cache)
	sort.Sort(batchedTransaction(transactions))
	for i := 0; i < len(transactions); i+=BATCH_READ_SIZE {
		go T.preloadCopaths(transactions[i:min(i+BATCH_READ_SIZE, len(transactions))], readChannel)
	}

	for i := 0; i < len(transactions); i+=BATCH_READ_SIZE {
		copathPairs := <-readChannel
		for j := 0; j < len(copathPairs); j++ {
			T.cache[copathPairs[j].ID] = &copathPairs[j].digest
		}
	}
	// fmt.Println("Cache after preload")
	// fmt.Println(T.cache)
	var err error
	T.conflicts, err = findConflicts(transactions)
	if err != nil {
		return false, err
	}

	for _, transaction := range transactions {
		for currID := transaction.id; currID != ""; currID = getParent(currID) {
			placeHolder := T.getEmpty(0)
			T.cache[currID] = &placeHolder
		}
	}
	placeHolder := T.getEmpty(0)
	T.cache[""] = &placeHolder

	var wg sync.WaitGroup
	ch := make(chan []*CoPathPair)
	quit := make(chan bool)

	go auroraWriteback(ch, quit)

	for i:=0; i<len(transactions); i++ {
		wg.Add(1)
		go T.percolate(transactions[i].id, transactions[i].data, &wg, ch)
	}
	wg.Wait()
	quit <- true
	rootDigestPointer := T.cache[""]
	T.rootDigest = *rootDigestPointer

	return true, nil
}

func (T *SparseMerkleTree) batch2Insert(transactions batchedTransaction) (bool, error) {
	readChannel := make(chan []*CoPathPair)
	sort.Sort(batchedTransaction(transactions))
	for i := 0; i < len(transactions); i+=BATCH_READ_SIZE {
		go T.preloadCopaths(transactions[i:min(i+BATCH_READ_SIZE, len(transactions))], readChannel)
	}

	for i := 0; i < len(transactions); i+=BATCH_READ_SIZE {
		copathPairs := <-readChannel
		for j := 0; j < len(copathPairs); j++ {
			T.cache[copathPairs[j].ID] = &copathPairs[j].digest
		}
	}

	var err error
	T.conflicts, err = findConflicts(transactions)
	if err != nil {
		return false, err
	}

	for _, transaction := range transactions {
		for currID := transaction.id; currID != ""; currID = getParent(currID) {
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

	go auroraWriteback(ch, quit)

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

func (T *SparseMerkleTree) percolate(index string, data string, wg *sync.WaitGroup, ch chan []*CoPathPair) (bool, error) {
	defer wg.Done()

	changeList := make([]*CoPathPair, 0)

	changeList = append(changeList, &CoPathPair{ID: index, digest: hashDigest([]byte(data))})
	
	//TODO: You should not hash the value passed in if it not a leaf ie in the root tree
	hash := hashDigest([]byte(data))
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
		changeList = append(changeList, &CoPathPair{ID: parentID, digest: parentDigest})
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

func (T *SparseMerkleTree) batchPercolate(transactions batchedTransaction, wg *sync.WaitGroup, ch chan []*CoPathPair) (bool, error) {
	defer wg.Done()
	changeList := make([]*CoPathPair, 0)
	
	for _, trans := range transactions {
		index := trans.id
		data := trans.data

		//TODO: You should not hash the value passed in if it not a leaf ie in the root tree
		hash := hashDigest([]byte(data))
		indexPointer := T.cache[index]
		*indexPointer = hash
		changeList = append(changeList, &CoPathPair{ID: index, digest: hash})

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
			changeList = append(changeList, &CoPathPair{ID: parentID, digest: parentDigest})

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

func (T *SparseMerkleTree) generateProof(index string) (Proof) {
	proofResult := Proof{}
	proofResult.queryID = index

	var proof_t ProofType
	var currID string
	_, ok := T.cache[index]
	if !ok {
		proof_t = NONMEMBERSHIP
		currID = T.getEmptyAncestor(currID)
	} else {
		proof_t = MEMBERSHIP
		currID = index
	}
	proofResult.proofType = proof_t
	proofResult.proofID = currID
	coPath := make([]CoPathPair, 0)

	// Our stopping condition is length > 0 so we don't add the root to the copath
	for ; len(currID) > 0; currID = getParent(currID) {
		// Append the sibling to the copath and advance current node
		siblingID, _ := getSibling(currID)
		siblingDigestPointer, ok := T.cache[siblingID]
		var siblingDigest digest
		if !ok {
			siblingDigest = T.getEmpty(len(siblingID))
		} else {
			siblingDigest = *siblingDigestPointer
		}

		coPathNode := CoPathPair{siblingID, siblingDigest}
		coPath = append(coPath, coPathNode)
	}

	proofResult.coPath = coPath
	return proofResult
}

func (T *SparseMerkleTree) verifyProof(proof Proof) (bool) {
	// If proof of nonmembership, first make sure that there is a prefix match
	proofIDLength := len(proof.proofID)
	if proof.proofType == NONMEMBERSHIP {
		if proofIDLength > len(proof.queryID) {
			return false
		}

		for i := 0; i < proofIDLength; i++ {
			if proof.proofID[i] != proof.queryID[i] {
				return false
			}
		}
	}

	rootDigestPointer, ok := T.cache[proof.proofID]
	var rootDigest digest
	if !ok {
		rootDigest = T.getEmpty(TREE_DEPTH - proofIDLength)
	} else {
		rootDigest = *rootDigestPointer
	}

	for i := 0; i < len(proof.coPath); i++ {
		currNode := proof.coPath[i]
		if currNode.ID[len(currNode.ID) - 1] == '0' {
			rootDigest = hashDigest(append(currNode.digest, rootDigest...))
		} else {
			rootDigest = hashDigest(append(rootDigest, currNode.digest...))
		}
	}
	return bytes.Equal(rootDigest, T.rootDigest)
}

// leaves must be sorted before findConflicts is called
func findConflicts(leaves []*transaction) (map[string]*SyncBool, error) {
	var conflicts = make(map[string]*SyncBool)
	for i := 1; i < len(leaves); i++ {
		x, y := leaves[i-1].id, leaves[i].id
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

func makeTree() (*SparseMerkleTree, error) {
	T := SparseMerkleTree{} 
	T.depth = TREE_DEPTH
	T.cache = make(map[string]*digest)
	T.empty_cache = make(map[int]digest)
	T.empty_cache[0] = hashDigest([]byte("0"))  
	T.rootDigest = T.getEmpty(TREE_DEPTH) // FIXME: Should this be the case for an empty tree?
	T.conflicts = make(map[string]*SyncBool)
	return &T, nil
}
