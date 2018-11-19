package merkle

import (
	_ "fmt"
	"hash"
	"bytes"
	"sync"
	"sort"
)

const TREE_DEPTH int = 256

type digest = []byte

type SparseMerkleTree struct {
	H hash.Hash 
	depth int
	cache map[string]digest
	rootDigest digest
	empty_cache map[int]digest
	conflicts map[string]*SyncBool
}

type SyncBool struct {
	lock *sync.Mutex
	writeable bool
}

func (T *SparseMerkleTree) getEmpty(n int) (digest) {
	if (len(T.empty_cache) <= n) {
		t := T.getEmpty(n - 1)
		T.empty_cache[n] = hashDigest(T.H, append(t[:], t[:]...))
	} 
	return T.empty_cache[n]
}

func hashDigest(H hash.Hash, data []byte) (digest) {
	defer H.Reset()
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
	// FIXME: What does this do exactly?
	currID := nodeID
	prevID := currID
	for (len(currID) > 0) {
		if _, contains := T.cache[currID]; contains {
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
	T.cache[index] = hashDigest(T.H, []byte(data))

	//FIXME: Actual copy?
	var currID string
	for currID = index; len(currID) > 0; {
		// Get both the parent and sibling IDs
		siblingID, isLeft := getSibling(currID)
		parentID := getParent(currID)

		// Get the digest of the current node and sibling
		currDigest := T.cache[currID]
		siblingDigest, contains := T.cache[siblingID] 
		if !contains {
			siblingDigest = T.getEmpty(len(siblingID))
		}

		// Hash the digests of the left and right children
		var parentDigest digest
		if isLeft {
			parentDigest = hashDigest(T.H, append(siblingDigest, currDigest...))
		} else {
			parentDigest = hashDigest(T.H, append(currDigest, siblingDigest...))
		}
		T.cache[parentID] = parentDigest

		// Traverse up the tree by making the current node the parent node
		currID = parentID
	}
	T.rootDigest = T.cache[currID]
	return true
}


func (T *SparseMerkleTree) batchInsert(transactions batchedTransaction) (bool, error) {

	sort.Sort(batchedTransaction(transactions))
	var err error
	T.conflicts, err = findConflicts(transactions)
	if err != nil {
		return false, err
	}

	var wg sync.WaitGroup

	for i:=0; i<len(transactions); i++ {
		wg.Add(1)
		//go T.percolate(transactions[i].id, transactions[i].data, wg)
	}
	wg.Wait()

	T.rootDigest = T.cache[""]

	return true, nil

}

func (T *SparseMerkleTree) generateProof(index string) (Proof) {
	proofResult := Proof{}
	proofResult.queryID = index

	var proof_t ProofType
	var currID string
	_, contains := T.cache[index]
	if !contains {
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
		siblingDigest, contains := T.cache[siblingID]
		if !contains {
			siblingDigest = T.getEmpty(len(siblingID))
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

	rootDigest, contains := T.cache[proof.proofID]
	if !contains {
		rootDigest = T.getEmpty(TREE_DEPTH - proofIDLength)
	}

	for i := 0; i < len(proof.coPath); i++ {
		currNode := proof.coPath[i]
		if currNode.ID[len(currNode.ID) - 1] == '0' {
			rootDigest = hashDigest(T.H, append(currNode.digest, rootDigest...))
		} else {
			rootDigest = hashDigest(T.H, append(rootDigest, currNode.digest...))
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
		conflicts[x[0:k]] = &SyncBool{lock: &sync.Mutex{}, writeable: false}
	}
	return conflicts, nil
}

func makeTree(H hash.Hash) (*SparseMerkleTree, error) {
	T := SparseMerkleTree{} 
	T.H = H
	T.depth = TREE_DEPTH
	T.cache = make(map[string]digest)	
	T.empty_cache = make(map[int]digest)
	T.empty_cache[0] = hashDigest(H, []byte("0"))  
	T.rootDigest = T.getEmpty(TREE_DEPTH) // FIXME: Should this be the case for an empty tree?
	T.conflicts = make(map[string]*SyncBool)
	return &T, nil
}
