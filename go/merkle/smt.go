package merkle

import (
	"hash"
	"math"
	"bytes"
	"math/big"
)

const TREE_DEPTH int = 128

type bitstring = big.Int
type digest = []byte

type SparseMerkleTree struct {
	hashName string
	depth int
	cache map[bitstring]digest
	empty_cache map[bitstring]digest
	root_digest digest
	conflicts map[bitstring]bool
}

func (T *SparseMerkleTree) 

func makeTree(hashName string)(*SparseMerkleTree, error) {
	T := SparseMerkleTree{hashName, TREE_DEPTH, leaves, H} 
}