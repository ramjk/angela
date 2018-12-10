package merkle

import (
	"encoding/base64"
	"fmt"
	"crypto/sha256"
	// "bytes"
	"sync"
	"strconv"
)

const TREE_DEPTH int = 256

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
func (T *SparseMerkleTree) Insert(index string, data string, epochNumber uint64) (bool) {
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

	for currID := index; currID != ""; currID = getParent(currID) {
		placeHolder := T.getEmpty(T.depth)
		T.cache[currID] = &placeHolder
	}
	placeHolder := T.getEmpty(T.depth)
	T.cache[""] = &placeHolder

	go T.preloadCopaths(BatchedTransaction{&Transaction{ID: index, Data: data}}, readChannel, readDB)

	copathPairs := <-readChannel
	for j := 0; j < len(copathPairs); j++ {
		T.cache[copathPairs[j].ID[len(T.prefix):]] = &copathPairs[j].Digest
	}

	ch := make(chan []*CoPathPair)
	quit := make(chan bool)

	go auroraWritebackBatch(ch, quit, writeDB, T.prefix, epochNumber, 1, 1)

	dig, _ := base64.StdEncoding.DecodeString(data)
	hash := hashDigest(dig)
	T.cache[index] = &hash

	//FIXME: Actual copy?
	var currID string
	changeList := make([]*CoPathPair, 0)
	changeList = append(changeList, &CoPathPair{ID: index, Digest: hashDigest(dig)})

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
		changeList = append(changeList, &CoPathPair{ID: parentID, Digest: parentDigest})

		// Traverse up the tree by making the current node the parent node
		currID = parentID
	}
	ch <- changeList
	quit <- true
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
    	// appending the subtree prefix to be read from the database
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

// Writes back to the database after receiving a set number of transactions
func auroraWritebackBatch(
	ch chan []*CoPathPair, 
	quit chan bool, 
	db *angelaDB, 
	prefix string, 
	epochNumber uint64, 
	batchWriteSize int,
	totalNumTransactions int) {

	bufferList := make([]*CoPathPair, 0)
	counter := 0
    for {
    	select {
    	case delta := <-ch:
    		// write back to aurora of the changelist from the channel
    		// fmt.Println("Changelist")
    		// fmt.Println(delta)
    		bufferList = append(bufferList, delta...)
    		counter += 1
    		if counter % batchWriteSize == 0 || counter == totalNumTransactions {
    			numRowsAffected, err := db.insertChangeList(prefix, bufferList, epochNumber)
	    		if err != nil {
	    			fmt.Println(err)
	    		}
	    		fmt.Println("Printing the number of rows affected", numRowsAffected)
	    		bufferList = nil
    		}
    		if counter == totalNumTransactions {
    			db.Close()
    		}
    	case <-quit:
    		// if len(bufferList) > 0 {
    		// 	numRowsAffected, err := db.insertChangeList(bufferList, epochNumber)
	    	// 	if err != nil {
	    	// 		fmt.Println(err)
	    	// 	}
	    	// 	fmt.Println("Printing the number of rows affected", numRowsAffected)		
    		// }
    		fmt.Println("Done Writing")
    		return
    	}
    }
}

func (T *SparseMerkleTree) BatchInsert(
	transactions BatchedTransaction, 
	epochNumber uint64, 
	batchReadSize int, 
	batchPercolateSize int, 
	batchWriteSize int) (string, error) {

	readChannel := make(chan []*CoPathPair)
	readDB, err := GetReadAngelaDB()
	if err != nil {
		panic(err)
	}
	defer readDB.Close()

	// Closing of this DB happens in auroraWriteback or when findConflicts errors
	writeDB, err := GetWriteAngelaDB()
	if err != nil {
		panic(err)
	}
	
	// fmt.Println("Cache before preload")
	// fmt.Println(T.cache)

	for _, Transaction := range transactions {
		for currID := Transaction.ID; currID != ""; currID = getParent(currID) {
			placeHolder := T.getEmpty(T.depth)
			T.cache[currID] = &placeHolder
		}
	}
	placeHolder := T.getEmpty(T.depth)
	T.cache[""] = &placeHolder

	for i := 0; i < len(transactions); i+=batchReadSize {
		go T.preloadCopaths(transactions[i:Min(i+batchReadSize, len(transactions))], readChannel, readDB)
	}

	for i := 0; i < len(transactions); i+=batchReadSize {
		copathPairs := <-readChannel
		for j := 0; j < len(copathPairs); j++ {
			// storing the read results in the cache by subtree virtual address
			T.cache[copathPairs[j].ID[len(T.prefix):]] = &copathPairs[j].Digest
		}
	}
	// fmt.Println("Cache after preload")
	// fmt.Println(T.cache)
	T.conflicts, err = findConflicts(transactions)
	if err != nil {
		writeDB.Close()
		return "", err
	}

	var wg sync.WaitGroup
	ch := make(chan []*CoPathPair)
	quit := make(chan bool)

	go auroraWritebackBatch(ch, quit, writeDB, T.prefix, epochNumber, batchWriteSize, len(transactions))

	for i:=0; i<len(transactions); i+=batchPercolateSize {
		wg.Add(1)
		go T.batchPercolate(transactions[i:Min(i+batchPercolateSize, len(transactions))], &wg, ch)
	}
	wg.Wait()
	quit <- true
	rootDigestPointer := T.cache[""]
	T.rootDigest = *rootDigestPointer

	return base64.StdEncoding.EncodeToString(T.rootDigest), nil
}

func (T *SparseMerkleTree) batchPercolate(transactions BatchedTransaction, wg *sync.WaitGroup, ch chan []*CoPathPair) (bool, error) {
	defer wg.Done()
	status := true
	for _, trans := range transactions {
		index := trans.ID
		data := trans.Data
		ok, err := T.percolate(index, data, ch)
		if err != nil {
			fmt.Println(err)
		}
		status = status && ok
	}
	return status, nil
}

func (T *SparseMerkleTree) percolate(index string, data string, ch chan []*CoPathPair) (bool, error) {
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

// Always getting the actual root of the tree
func (T *SparseMerkleTree) GetLatestRoot() (string) {
	rootDigest, _ := T.getLatestNode("")
	return base64.StdEncoding.EncodeToString(rootDigest)
}

func (T *SparseMerkleTree) getLatestNode(nodeId string) (digest, bool) {
	readDB, err := GetReadAngelaDB()
	if err != nil {
		panic(err)
	}
	defer readDB.Close()
	nodeIds := []string{nodeId}
	copathPairs, err := readDB.retrieveLatestCopathDigests(nodeIds)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("")
	ok := err == nil && len(copathPairs) == 1
	if ok {
		return copathPairs[0].Digest, ok
	}
	return T.getEmpty(T.depth), ok
}

func (T *SparseMerkleTree) CGenerateProof(index string) ([]string) {
	proof := T.GenerateProofDB(index)
	results := make([]string, proof.ProofLength)
	fmt.Println("About to format ",bool(proof.ProofType))
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

// IDs dealt with here should be non-virtual (i.e. no prefix involved)
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
	_, ok := T.getLatestNode(index)

	// _, ok := T.cache[index]
	if !ok {
		// fmt.Println("Entering NONMEMBERSHIP")
		proof_t = NONMEMBERSHIP
		ancestorIds := make([]string, 0)
		ancestor := index
		// Our stopping condition is length > 0 so we don't add the root to the copath
		for ; len(ancestor) > 0; ancestor = getParent(ancestor) {
			ancestorIds = append(ancestorIds, ancestor)
		}
	    // fmt.Println("number of ancestors", len(ancestorIds))

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
	// fmt.Println("startingIndex is", startingIndex)
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
	// fmt.Println("Length of CoPath", len(CoPath))
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
		// fmt.Println("verifying proof of nonmembership")
		if ProofIDLength > len(proof.QueryID) {
			return false
		}

		for i := 0; i < ProofIDLength; i++ {
			if proof.ProofID[i] != proof.QueryID[i] {
				return false
			}
		}
	}
	var rootDigest digest
	rootDigest, ok := T.getLatestNode(proof.ProofID) 
	if !ok {
		// fmt.Println("not ok rootDigest verify")
		rootDigest = T.getEmpty(T.depth - ProofIDLength)
	}

	for i := 0; i < len(proof.CoPath); i++ {
		currNode := proof.CoPath[i]
		if currNode.ID[len(currNode.ID) - 1] == '0' {
			rootDigest = hashDigest(append(currNode.Digest, rootDigest...))
		} else {
			rootDigest = hashDigest(append(rootDigest, currNode.Digest...))
		}
	}
	good := base64.StdEncoding.EncodeToString(rootDigest) == T.GetLatestRoot()
	if !good {
		fmt.Println("Root Digest Calculated", base64.StdEncoding.EncodeToString(rootDigest))
		fmt.Println("Latest Root", T.GetLatestRoot())
		fmt.Println("ProofIDLength", ProofIDLength)
		fmt.Println("Length of copath", len(proof.CoPath))
		fmt.Println("first copath item id", proof.CoPath[0].ID)
		fmt.Println("first copath item digest", proof.CoPath[0].Digest)
		fmt.Println("last copath item id", proof.CoPath[255].ID)
		fmt.Println("last copath item digest", proof.CoPath[255].Digest)
	}
	return good
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
