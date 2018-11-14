package merkle

import (
	"time"
	// "fmt"
	"testing"
	"bytes"
	"crypto/sha256"
	"math/rand"
)

const NUMITERATIONS int = 1000

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

func TestConstructor(t *testing.T) {
	SHA256 := sha256.New()
	tree, _ := makeTree(SHA256)

	actual := tree.getEmpty(0)
	expected := hashDigest(SHA256, []byte("0"))
	if !bytes.Equal(actual, expected) {
		t.Error("0-th level empty node is incorrect.")
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
	// indices := make([]string, 1)
	// for i := 0; i < NUMITERATIONS; i++ {
	// 	append(indices, randomBitString(128))
	// }
}