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

const NUMITERATIONS int = 10000

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

	index := randomBitString(TREE_DEPTH)

	tree.insert(index, "angela")

	proof := tree.generateProof(index)

	if len(proof.coPath) != TREE_DEPTH {
		t.Error("Length of the copath was not equal to TREE_DEPTH.")
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
		indices = append(indices, randomBitString(TREE_DEPTH))
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
		if len(proof.coPath) != TREE_DEPTH {
			t.Error("Length of copath != TREE_DEPTH")
		}
		if proof.proofType == NONMEMBERSHIP {
			t.Error("Proof of non-membership")
		}
		if !tree.verifyProof(proof) {
			t.Error("Proof was invalid when it was expected to be valid.")
		}
	}
}

func benchmarkInsertN(tree *SparseMerkleTree, indices []string, data []string, b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i, index := range indices {
			tree.insert(index, data[i])
		}
	}
}

func BenchmarkInsert64(b *testing.B) {
	SHA256 := sha256.New()
	tree, _ := makeTree(SHA256)

	indices := make([]string, 64)
	data := make([]string, 64)

	for i := 0; i < 64; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert128(b *testing.B) {
	SHA256 := sha256.New()
	tree, _ := makeTree(SHA256)

	indices := make([]string, 128)
	data := make([]string, 128)

	for i := 0; i < 128; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert256(b *testing.B) {
	SHA256 := sha256.New()
	tree, _ := makeTree(SHA256)

	indices := make([]string, 256)
	data := make([]string, 256)

	for i := 0; i < 256; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert512(b *testing.B) {
	SHA256 := sha256.New()
	tree, _ := makeTree(SHA256)

	indices := make([]string, 512)
	data := make([]string, 512)

	for i := 0; i < 512; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)	
}

func BenchmarkInsert1024(b *testing.B) {
	SHA256 := sha256.New()
	tree, _ := makeTree(SHA256)

	indices := make([]string, 1024)
	data := make([]string, 1024)

	for i := 0; i < 1024; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert2048(b *testing.B) {
	SHA256 := sha256.New()
	tree, _ := makeTree(SHA256)

	indices := make([]string, 2048)
	data := make([]string, 2048)

	for i := 0; i < 2048; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert4096(b *testing.B) {
	SHA256 := sha256.New()
	tree, _ := makeTree(SHA256)

	indices := make([]string, 4096)
	data := make([]string, 4096)

	for i := 0; i < 4096; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert8096(b *testing.B) {
	SHA256 := sha256.New()
	tree, _ := makeTree(SHA256)

	indices := make([]string, 8096)
	data := make([]string, 8096)

	for i := 0; i < 8096; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}
