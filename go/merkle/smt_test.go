package merkle

import (
	"fmt"
	"testing"
	"bytes"
	"crypto/sha256"
)

func TestConstructor(t *testing.T) {
	SHA256 := sha256.New()
	tree, _ := makeTree(SHA256)

	fmt.Println(len(tree.empty_cache))

	actual := tree.getEmpty(0)
	expected := hashDigest(SHA256, []byte("0"))
	if !bytes.Equal(actual, expected) {
		t.Error("0-th level empty node is incorrect.")
	}
}
