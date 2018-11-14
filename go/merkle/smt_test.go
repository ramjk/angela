package merkle

import (
	_ "fmt"
	"testing"
	"bytes"
	"crypto/sha256"
)

func TestConstructor(t *testing.T) {
	SHA256 := sha256.New()
	tree, _ := makeTree(SHA256)

	actual := tree.getEmpty(0)
	expected := hashDigest(SHA256, []byte("0"))
	if !bytes.Equal(actual, expected) {
		t.Error("0-th level empty node is incorrect.")
	}
}

func TestMemberShipSmall(t *testing.T) {
	index := "101"

	SHA256 := sha256.New()
	tree, _ := makeTree(SHA256)

	tree.insert(index, "angela")

	proof := tree.generateProof(index)

	if !tree.verifyProof(proof) {
		t.Error("Proof was invalid when it was expected to be valid.")
	}
}
