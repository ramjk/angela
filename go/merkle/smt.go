package merkle

import (
	"hash"
	"crypto/sha256"
	"fmt"
)

const TREE_DEPTH int = 128

type digest = []byte

type SparseMerkleTree struct {
	H hash.Hash 
	depth int
	cache map[string]digest
	root_digest digest
	empty_cache map[int]digest
	conflicts map[string]bool
}

func makeTree(H hash.Hash) (*SparseMerkleTree, error) {
	T := SparseMerkleTree{} 
	T.H = H
	T.depth = TREE_DEPTH
	T.cache = make(map[string]digest)
	T.root_digest = hashDigest(H, []byte("0")) // FIXME: Should this be the case for an empty tree?
	T.empty_cache = make(map[int]digest)
	T.empty_cache[0] = T.root_digest 
	T.conflicts = make(map[string]bool)
	return &T, nil
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
		return nodeID
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

	return siblingID
}

func getEmptyAncestor(nodeID string) {
	// FIXME: What does this do exactly?
}

/* Assume index is valid (for now).

FIXME: What are the error cases where we return error/False?
*/
func (T *SparseMerkleTree) insert(index string, data string) (bool) {
	T.cache[index] = hashDigest(data)

	//FIXME: Actual copy?
	var currID string
	for (currID = index; len(currID) > 0;) {
		// Get both the parent and sibling IDs
		siblingID, isLeft := getSibling(currID)
		parentID := getParent(currID)

		// Get the digest of the current node and sibling
		currDigest := T.cache[currID]
		if siblingDigest, err := T.cache[siblingID]; err {
			siblingDigest = getEmpty(len(siblingID))
		}

		// Hash the digests of the left and right children
		var parentDigest digest
		if isLeft {
			parentDigest = hashDigest(T.H, siblingDigest + currDigest)
		} else {
			parentDigest = hashDigest(T.H, currDigest + siblingDigest)
		}
		T.cache[parentID] = parentDigest

		// Traverse up the tree by making the current node the parent node
		currentID = parentID
	}

	T.root_digest = T.cache[currID]
	return true
}

func (T *SparseMerkleTree) generateProof(index string) (Proof) {
	proofResult := Proof{}
	proofResult.queryID = index
	coPath := make([]CoPathPair)

	var proof_t ProofType
	if currID, notContains := T.cache[index]; notContains {
		proof_t = NONMEMBERSHIP
		currID = getEmptyAncestor(currID)
	} else {
		proof_t = MEMBERSHIP
	}
	proofResult.proofID = currID

	// Our stopping condition is length > 0 so we don't add the root to the copath
	for (; len(currID) > 0; currID = getParent(currID)) {
		// Append the sibling to the copath and advance current node
		siblingID, isLeft := getSibling(currID)
		if siblingDigest, err := T.cache[siblingID]; err {
			siblingDigest = getEmpty(len(siblingID))
		}

		coPathNode := CoPathPair{siblingID, siblingDigest}
		coPath += coPathNode
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

		for i, _ := range [proofIDLength]int {
			if proof.proofID[i] != proof.queryID[i] {
				return false
			}
		}
	}
}

func main() {
	H := sha256.New()
	T, _ := makeTree(H)
	fmt.Println(T.depth)
	myBits := "01111000"
	fmt.Println("myBits")
	fmt.Println(myBits)
	parent := getParent(myBits)
	fmt.Println("parent")
	fmt.Println(parent)
	sibling := getSibling(myBits)
	fmt.Println(sibling)
	T.cache["01111001"] = hashDigest(T.H, []byte(myBits))
	fmt.Println(T.cache["01111001"])
}