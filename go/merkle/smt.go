package main

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

func getSibling(nodeID string) (string) {
	length := len(nodeID)
	if (length == 0) {
		return nodeID
	}

	
	lastBit := nodeID[length - 1]
	siblingID := nodeID[:length - 1]

	if (lastBit == byte('0')) {
		siblingID += "1"
	} else {
		siblingID += "0"
	}

	return siblingID
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