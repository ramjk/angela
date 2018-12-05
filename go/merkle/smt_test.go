package merkle

import (
	"time"
	"fmt"
	"testing"
	"strings"
	"bytes"
	"math/rand"
	"sync"
	"sort"
)

const NUMITERATIONS int = 100

// var seedNum int64 = 0

func randomBitString(digestSize int) (string) {
	rand.Seed(time.Now().UTC().UnixNano())
	// use to get same input over multiple runs
	// rand.Seed(seedNum)
	// seedNum += 1
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

func TestFindConflicts(t *testing.T) {
	leaves := []*Transaction{{"0000", "data"}, {"0001", "data"}, 
							 {"0010", "data"}, {"0110", "data"}, 
							 {"1110", "data"}} 
	var correct_conflict = make(map[string]*SyncBool)
	correct_conflict[""] = &SyncBool{&sync.Mutex{}, false}
	correct_conflict["0"] = &SyncBool{&sync.Mutex{}, false}
	correct_conflict["00"] = &SyncBool{&sync.Mutex{}, false}		
	correct_conflict["000"] = &SyncBool{&sync.Mutex{}, false}		
	
	conflicts, err := findConflicts(leaves)

	if err != nil || len(conflicts) != len(correct_conflict) {
		t.Error("Incorrect number of conflicts generated")
	}

	for k, _ := range correct_conflict {
		val, ok := conflicts[k];
		if !ok {
		    t.Error("Conflict not found:")
		}
		if val.visited {
			t.Error("visited should be false, is true")
		}	
	}
}

func TestSortTransactions(t *testing.T) {
	arrLen := 10

	Transactions := make([]*Transaction, arrLen)

	for i:=0; i < arrLen; i++ {

		Transactions[i] = &Transaction{randomBitString(128), "data"}
	}

	sort.Sort(batchedTransaction(Transactions))

	for i:=0; i < arrLen - 1; i++ {
		if Transactions[i].ID > Transactions[i+1].ID {
			t.Error("SORT is broken")
		}
	}
}

// func TestRunDBCommands(t *testing.T) {
// 	db, err := GetWriteAngelaDB()
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer db.Close()
// 	db.ShowTables()
// 	db.CreateTable()
// 	db.ShowTables() 
// }

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
	tree := MakeTree()

	if !bytes.Equal(tree.empty_cache[0], tree.getEmpty(0)) {
		t.Error("empty_cache[0] != getEmpty(0)")
	}

	actual := tree.getEmpty(0)
	expected := hashDigest([]byte(""))
	if !bytes.Equal(actual, expected) {
		t.Error("0-th level empty node is incorrect.")
	}

	if !bytes.Equal(tree.getEmpty(TREE_DEPTH), tree.rootDigest) {
		t.Error("Root Digest was not equal to getEmpty(TREE_DEPTH)")
	}
}

func TestMembershipSmall(t *testing.T) {
	index := "101"

	tree := MakeTree()

	tree.Insert(index, "angela")

	proof := tree.GenerateProof(index)

	if !tree.verifyProof(proof) {
		t.Error("Proof was invalid when it was expected to be valid.")
	}
}

func TestMembership(t *testing.T) {
	tree := MakeTree()

	index := randomBitString(TREE_DEPTH)

	tree.Insert(index, "angela")

	proof := tree.GenerateProof(index)

	if len(proof.CoPath) != TREE_DEPTH {
		t.Error("Length of the copath was not equal to TREE_DEPTH.")
	}

	if !tree.verifyProof(proof) {
		t.Error("Proof was invalid when it was expected to be valid.")
	}
}

func TestMembershipLarge(t *testing.T) {
	tree := MakeTree()

	indices := make([]string, 0)
	for i := 0; i < NUMITERATIONS; i++ {
		indices = append(indices, randomBitString(TREE_DEPTH))
	}

	for i, bitString := range indices {
		data := fmt.Sprintf("angela%d", i)
		tree.Insert(bitString, data)
	}

	proofs := make([]Proof, len(indices))

	for i, bitString := range indices {
		proofs[i] = tree.GenerateProof(bitString)
	}

	for i, proof := range proofs {
		if strings.Compare(proof.ProofID, proof.QueryID) != 0 {
			t.Error("proofID != queryID")
		}
		if strings.Compare(proof.ProofID, indices[i]) != 0 {
			t.Error("i-th proofID != indices[i]")
		}
		if len(proof.CoPath) != len(proof.ProofID) {
			t.Error("Length of coPath != proofID")
		}
		if len(proof.CoPath) != TREE_DEPTH {
			t.Error("Length of copath != TREE_DEPTH")
		}
		if proof.ProofType == NONMEMBERSHIP {
			t.Error("Proof of non-membership")
		}
		if !tree.verifyProof(proof) {
			t.Error("Proof was invalid when it was expected to be valid.")
		}
	}
}

func TestNonMembership(t *testing.T) {
	tree := MakeTree()

	queryID := randomBitString(128)
	proof := tree.GenerateProof(queryID)

	if proof.ProofType == MEMBERSHIP {
		t.Error("Proof should be of type nonmembership")
	}

	if tree.verifyProof(proof) == false {
		t.Error("Proof was not verified")
	}
}


func benchmarkInsertN(tree *SparseMerkleTree, indices []string, data []string, b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i, index := range indices {
			tree.Insert(index, data[i])
		}
	}
}

func BenchmarkInsert64(b *testing.B) {
	tree := MakeTree()

	indices := make([]string, 64)
	data := make([]string, 64)

	for i := 0; i < 64; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert128(b *testing.B) {
	tree := MakeTree()

	indices := make([]string, 128)
	data := make([]string, 128)

	for i := 0; i < 128; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert256(b *testing.B) {
	tree := MakeTree()

	indices := make([]string, 256)
	data := make([]string, 256)

	for i := 0; i < 256; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert512(b *testing.B) {
	tree := MakeTree()

	indices := make([]string, 512)
	data := make([]string, 512)

	for i := 0; i < 512; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)	
}

func BenchmarkInsert1024(b *testing.B) {
	tree := MakeTree()

	indices := make([]string, 1024)
	data := make([]string, 1024)

	for i := 0; i < 1024; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert2048(b *testing.B) {
	tree := MakeTree()

	indices := make([]string, 2048)
	data := make([]string, 2048)

	for i := 0; i < 2048; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert4096(b *testing.B) {
	tree := MakeTree()

	indices := make([]string, 4096)
	data := make([]string, 4096)

	for i := 0; i < 4096; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert8192(b *testing.B) {
	tree := MakeTree()

	indices := make([]string, 8192)
	data := make([]string, 8192)

	for i := 0; i < 8192; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert16384(b *testing.B) {
	tree := MakeTree()

	indices := make([]string, 16384)
	data := make([]string, 16384)

	for i := 0; i < 16384; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkBatchInsert64(b *testing.B) {
	tree := MakeTree()

	Transactions := make([]*Transaction, 64)

	for i := 0; i < 64; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := Transaction{ID: index, Data: d}
		Transactions[i] = &t
	}

	b.ResetTimer()
	tree.BatchInsert(Transactions)
}

func BenchmarkBatchInsert128(b *testing.B) {
	tree := MakeTree()

	Transactions := make([]*Transaction, 128)

	for i := 0; i < 128; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := Transaction{ID: index, Data: d}
		Transactions[i] = &t
	}

	b.ResetTimer()
	tree.BatchInsert(Transactions)
}

func BenchmarkBatchInsert256(b *testing.B) {
	tree := MakeTree()

	Transactions := make([]*Transaction, 256)

	for i := 0; i < 256; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := Transaction{ID: index, Data: d}
		Transactions[i] = &t
	}

	b.ResetTimer()
	tree.BatchInsert(Transactions)
}

func BenchmarkBatchInsert512(b *testing.B) {
	tree := MakeTree()

	Transactions := make([]*Transaction, 512)

	for i := 0; i < 512; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := Transaction{ID: index, Data: d}
		Transactions[i] = &t
	}

	b.ResetTimer()
	tree.BatchInsert(Transactions)
}

func BenchmarkBatchInsert1024(b *testing.B) {
	tree := MakeTree()

	Transactions := make([]*Transaction, 1024)

	for i := 0; i < 1024; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := Transaction{ID: index, Data: d}
		Transactions[i] = &t
	}

	b.ResetTimer()
	tree.BatchInsert(Transactions)
}

func BenchmarkBatchInsert2048(b *testing.B) {
	tree := MakeTree()

	Transactions := make([]*Transaction, 2048)

	for i := 0; i < 2048; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := Transaction{ID: index, Data: d}
		Transactions[i] = &t
	}

	b.ResetTimer()
	tree.BatchInsert(Transactions)
}

func BenchmarkBatchInsert4096(b *testing.B) {
	tree := MakeTree()

	Transactions := make([]*Transaction, 4096)

	for i := 0; i < 4096; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := Transaction{ID: index, Data: d}
		Transactions[i] = &t
	}

	b.ResetTimer()
	tree.BatchInsert(Transactions)
}

func BenchmarkBatchInsert8192(b *testing.B) {
	tree := MakeTree()

	Transactions := make([]*Transaction, 8192)

	for i := 0; i < 8192; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := Transaction{ID: index, Data: d}
		Transactions[i] = &t
	}

	b.ResetTimer()
	tree.BatchInsert(Transactions)
}

func BenchmarkBatchInsert16384(b *testing.B) {
	tree := MakeTree()

	Transactions := make([]*Transaction, 16384)

	for i := 0; i < 16384; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := Transaction{ID: index, Data: d}
		Transactions[i] = &t
	}

	b.ResetTimer()
	tree.BatchInsert(Transactions)
}

func BenchmarkBatch2Insert64(b *testing.B) {
	tree := MakeTree()

	Transactions := make([]*Transaction, 64)

	for i := 0; i < 64; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := Transaction{ID: index, Data: d}
		Transactions[i] = &t
	}

	b.ResetTimer()
	tree.batch2Insert(Transactions)
}

func BenchmarkBatch2Insert128(b *testing.B) {
	tree := MakeTree()

	Transactions := make([]*Transaction, 128)

	for i := 0; i < 128; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := Transaction{ID: index, Data: d}
		Transactions[i] = &t
	}

	b.ResetTimer()
	tree.batch2Insert(Transactions)
}

func BenchmarkBatch2Insert256(b *testing.B) {
	tree := MakeTree()

	Transactions := make([]*Transaction, 256)

	for i := 0; i < 256; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := Transaction{ID: index, Data: d}
		Transactions[i] = &t
	}

	b.ResetTimer()
	tree.batch2Insert(Transactions)
}

func BenchmarkBatch2Insert512(b *testing.B) {
	tree := MakeTree()

	Transactions := make([]*Transaction, 512)

	for i := 0; i < 512; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := Transaction{ID: index, Data: d}
		Transactions[i] = &t
	}

	b.ResetTimer()
	tree.batch2Insert(Transactions)
}

func BenchmarkBatch2Insert1024(b *testing.B) {
	tree := MakeTree()

	Transactions := make([]*Transaction, 1024)

	for i := 0; i < 1024; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := Transaction{ID: index, Data: d}
		Transactions[i] = &t
	}

	b.ResetTimer()
	tree.batch2Insert(Transactions)
}

func BenchmarkBatch2Insert2048(b *testing.B) {
	tree := MakeTree()

	Transactions := make([]*Transaction, 2048)

	for i := 0; i < 2048; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := Transaction{ID: index, Data: d}
		Transactions[i] = &t
	}

	b.ResetTimer()
	tree.batch2Insert(Transactions)
}

func BenchmarkBatch2Insert4096(b *testing.B) {
	tree := MakeTree()

	Transactions := make([]*Transaction, 4096)

	for i := 0; i < 4096; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := Transaction{ID: index, Data: d}
		Transactions[i] = &t
	}

	b.ResetTimer()
	tree.batch2Insert(Transactions)
}

func BenchmarkBatch2Insert8192(b *testing.B) {
	tree := MakeTree()

	Transactions := make([]*Transaction, 8192)

	for i := 0; i < 8192; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := Transaction{ID: index, Data: d}
		Transactions[i] = &t
	}

	b.ResetTimer()
	tree.batch2Insert(Transactions)
}

func BenchmarkBatch2Insert16384(b *testing.B) {
	tree := MakeTree()

	Transactions := make([]*Transaction, 16384)

	for i := 0; i < 16384; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := Transaction{ID: index, Data: d}
		Transactions[i] = &t
	}

	b.ResetTimer()
	tree.batch2Insert(Transactions)
}

func TestBatchInsert(t * testing.T) {
	TransactionLen := NUMITERATIONS
	tree := MakeTree()

	Transactions := make([]*Transaction, TransactionLen)

	for i := 0; i < TransactionLen; i++ {
		Transactions[i] = &Transaction{randomBitString(TREE_DEPTH), fmt.Sprintf("angela%d", i)}
	}

    tree.BatchInsert(Transactions)
	// for k, v := range tree.conflicts { 
    //   fmt.Printf("key[%s] value[%s]\n", k, v.writeable)
	// }
	
	for i := 0; i < TransactionLen; i++ {
		proof := tree.GenerateProof(Transactions[i].ID)

		if len(proof.CoPath) != TREE_DEPTH {
			t.Error("Length of the copath was not equal to TREE_DEPTH.")
		}

		if !tree.verifyProof(proof) {
			t.Error("Proof was invalid when it was expected to be valid.")
		}
	}
}

func TestBatch2Insert(t * testing.T) {
	TransactionLen := NUMITERATIONS
	tree := MakeTree()

	Transactions := make([]*Transaction, TransactionLen)

	for i := 0; i < TransactionLen; i++ {
		Transactions[i] = &Transaction{randomBitString(TREE_DEPTH), fmt.Sprintf("angela%d", i)}
	}

    tree.batch2Insert(Transactions)
	// for k, v := range tree.conflicts { 
    //   fmt.Printf("key[%s] value[%s]\n", k, v.writeable)
	// }
	
	for i := 0; i < TransactionLen; i++ {
		proof := tree.GenerateProof(Transactions[i].ID)

		if len(proof.CoPath) != TREE_DEPTH {
			t.Error("Length of the copath was not equal to TREE_DEPTH.")
		}

		if !tree.verifyProof(proof) {
			t.Error("Proof was invalid when it was expected to be valid.")
		}
	}
}

func TestDatabaseConnection(t *testing.T) {
	db, err := GetReadAngelaDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()
	// Write a ping here 
}