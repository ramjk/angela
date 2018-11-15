package merkle

import (
	"time"
	"fmt"
	"testing"
	"strings"
	"bytes"
	"crypto/sha256"
	"math/rand"
)

const NUMITERATIONS int = 2

func randomBitString(digestSize int) (string) {
	rand.Seed(time.Now().UTC().UnixNano())
	result := ""
	for i := 0; i < digestSize; i++ {
		bit := rand.Int31n(2)
		if bit == 0 {
			result += "0"
		} else {
			result += "1"
		}
	}
	return result
}

func TestSibling(t *testing.T) {
	index := "11100010"
	_, isLeft := getSibling(index)
	if isLeft {
		t.Error("Sibling is not the left sibling")
	}
}

func TestParentEmpty(t *testing.T) {
	emptyZero := "0"
	emptyOne := "1"

	parentZero := getParent(emptyZero)
	parentOne := getParent(emptyOne)

	if strings.Compare(parentZero, parentOne) != 0 {
		t.Error("Parents of level 0 children were not equal")
	}

	if strings.Compare(parentZero, "") != 0 {
		t.Error("Parent of level 0 child is invalid")
	}
}

func TestConstructor(t *testing.T) {
	SHA256 := sha256.New()
	tree, _ := makeTree(SHA256)

	if !bytes.Equal(tree.empty_cache[0], tree.getEmpty(0)) {
		t.Error("empty_cache[0] != getEmpty(0)")
	}

	actual := tree.getEmpty(0)
	expected := hashDigest(SHA256, []byte("0"))
	if !bytes.Equal(actual, expected) {
		t.Error("0-th level empty node is incorrect.")
	}

	if !bytes.Equal(tree.getEmpty(TREE_DEPTH), tree.rootDigest) {
		t.Error("Root Digest was not equal to getEmpty(TREE_DEPTH)")
	}
}

func TestMembershipSmall(t *testing.T) {
	index := "101"

	SHA256 := sha256.New()
	tree, _ := makeTree(SHA256)

	tree.insert(index, "angela")

	proof := tree.generateProof(index)

	if !tree.verifyProof(proof) {
		t.Error("Proof was invalid when it was expected to be valid.")
	}
}

func TestMembership(t *testing.T) {
	SHA256 := sha256.New()
	tree, _ := makeTree(SHA256)

	index := randomBitString(128)

	tree.insert(index, "angela")

	proof := tree.generateProof(index)

	if len(proof.coPath) != 128 {
		t.Error("Length of the copath was not equal to 128.")
	}

	if !tree.verifyProof(proof) {
		t.Error("Proof was invalid when it was expected to be valid.")
	}
}

func TestMembershipLarge(t *testing.T) {
	SHA256 := sha256.New()
	tree, _ := makeTree(SHA256)

	indices := make([]string, 0)
	for i := 0; i < NUMITERATIONS; i++ {
		indices = append(indices, randomBitString(128))
	}

	for i, bitString := range indices {
		data := fmt.Sprintf("angela%d", i)
		tree.insert(bitString, data)
	}

	proofs := make([]Proof, len(indices))

	for i, bitString := range indices {
		proofs[i] = tree.generateProof(bitString)
	}

	for i, proof := range proofs {
		if strings.Compare(proof.proofID, proof.queryID) != 0 {
			t.Error("proofID != queryID")
		}
		if strings.Compare(proof.proofID, indices[i]) != 0 {
			t.Error("i-th proofID != indices[i]")
		}
		if len(proof.coPath) != len(proof.proofID) {
			t.Error("Length of coPath != proofID")
		}
		if len(proof.coPath) != 128 {
			t.Error("Length of copath != 128")
		}
		if proof.proofType == NONMEMBERSHIP {
			t.Error("Proof of non-membership")
		}
		if !tree.verifyProof(proof) {
			t.Error("Proof was invalid when it was expected to be valid.")
		}
	}
}