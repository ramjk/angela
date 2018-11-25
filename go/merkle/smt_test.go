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

const NUMITERATIONS int = 150

var seedNum int64 = 0

func randomBitString(digestSize int) (string) {
	rand.Seed(time.Now().UTC().UnixNano())
	// use to get same input over multiple runs
	rand.Seed(seedNum)
	seedNum += 1
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
	leaves := []*transaction{{"0000", "data"}, {"0001", "data"}, 
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

	transactions := make([]*transaction, arrLen)

	for i:=0; i < arrLen; i++ {

		transactions[i] = &transaction{randomBitString(128), "data"}
	}

	sort.Sort(batchedTransaction(transactions))

	for i:=0; i < arrLen - 1; i++ {
		if transactions[i].id > transactions[i+1].id {
			t.Error("SORT is broken")
		}
	}
}

func TestBatchInsert(t * testing.T) {
	transactionLen := NUMITERATIONS
	tree, _ := makeTree()

	transactions := make([]*transaction, transactionLen)

	for i := 0; i < transactionLen; i++ {
		transactions[i] = &transaction{randomBitString(TREE_DEPTH), fmt.Sprintf("angela%d", i)}
	}

	_ = tree

    tree.batchInsert(transactions)
	
	for i := 0; i < transactionLen; i++ {
		proof := tree.generateProof(transactions[i].id)

		if len(proof.coPath) != TREE_DEPTH {
			t.Error("Length of the copath was not equal to TREE_DEPTH.")
		}

		if !tree.verifyProof(proof) {
			t.Error("Proof was invalid when it was expected to be valid.")
		}
	}
}

func TestBatch2Insert(t * testing.T) {
	transactionLen := NUMITERATIONS
	tree, _ := makeTree()

	transactions := make([]*transaction, transactionLen)

	for i := 0; i < transactionLen; i++ {
		transactions[i] = &transaction{randomBitString(TREE_DEPTH), fmt.Sprintf("angela%d", i)}
	}

	_ = tree

    tree.batch2Insert(transactions)

	for i := 0; i < transactionLen; i++ {
		proof := tree.generateProof(transactions[i].id)

		if len(proof.coPath) != TREE_DEPTH {
			t.Error("Length of the copath was not equal to TREE_DEPTH.")
		}

		if !tree.verifyProof(proof) {
			t.Error("Proof was invalid when it was expected to be valid.")
		}
	}
}

func TestGetLastThousandNodes(t *testing.T) {
	nodes, err := getLastThousandNodes()
	if err != nil {
		t.Error(err)
	}
	for _, elem := range nodes {
		fmt.Println("NODE")
		fmt.Println(elem.ID)
		fmt.Println(elem.digest)
	}
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
	tree, _ := makeTree()

	if !bytes.Equal(tree.empty_cache[0], tree.getEmpty(0)) {
		t.Error("empty_cache[0] != getEmpty(0)")
	}

	actual := tree.getEmpty(0)
	expected := hashDigest([]byte("0"))
	if !bytes.Equal(actual, expected) {
		t.Error("0-th level empty node is incorrect.")
	}

	if !bytes.Equal(tree.getEmpty(TREE_DEPTH), tree.rootDigest) {
		t.Error("Root Digest was not equal to getEmpty(TREE_DEPTH)")
	}
}

func TestMembershipSmall(t *testing.T) {
	index := "101"

	tree, _ := makeTree()

	tree.insert(index, "angela")

	proof := tree.generateProof(index)

	if !tree.verifyProof(proof) {
		t.Error("Proof was invalid when it was expected to be valid.")
	}
}

func TestMembership(t *testing.T) {
	tree, _ := makeTree()

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
	tree, _ := makeTree()

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

func TestNonMembership(t *testing.T) {
	tree, _ := makeTree()

	queryID := randomBitString(128)
	proof := tree.generateProof(queryID)

	if proof.proofType == MEMBERSHIP {
		t.Error("Proof should be of type nonmembership")
	}

	if tree.verifyProof(proof) == false {
		t.Error("Proof was not verified")
	}
}

func TestDatabaseConnection(t *testing.T) {
	db, err := getAngelaDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	id, err := db.insertNode("001", "RAM", 1)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)

	nodeDigest, err := db.getLatestNodeDigest("001", 2)
	if err != nil {
		panic(err)
	}
	fmt.Println(nodeDigest)
}



func benchmarkInsertN(tree *SparseMerkleTree, indices []string, data []string, b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i, index := range indices {
			tree.insert(index, data[i])
		}
	}
}

func BenchmarkInsert64(b *testing.B) {
	tree, _ := makeTree()

	indices := make([]string, 64)
	data := make([]string, 64)

	for i := 0; i < 64; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert128(b *testing.B) {
	tree, _ := makeTree()

	indices := make([]string, 128)
	data := make([]string, 128)

	for i := 0; i < 128; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert256(b *testing.B) {
	tree, _ := makeTree()

	indices := make([]string, 256)
	data := make([]string, 256)

	for i := 0; i < 256; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert512(b *testing.B) {
	tree, _ := makeTree()

	indices := make([]string, 512)
	data := make([]string, 512)

	for i := 0; i < 512; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)	
}

func BenchmarkInsert1024(b *testing.B) {
	tree, _ := makeTree()

	indices := make([]string, 1024)
	data := make([]string, 1024)

	for i := 0; i < 1024; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert2048(b *testing.B) {
	tree, _ := makeTree()

	indices := make([]string, 2048)
	data := make([]string, 2048)

	for i := 0; i < 2048; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert4096(b *testing.B) {
	tree, _ := makeTree()

	indices := make([]string, 4096)
	data := make([]string, 4096)

	for i := 0; i < 4096; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert8192(b *testing.B) {
	tree, _ := makeTree()

	indices := make([]string, 8192)
	data := make([]string, 8192)

	for i := 0; i < 8192; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkInsert16384(b *testing.B) {
	tree, _ := makeTree()

	indices := make([]string, 16384)
	data := make([]string, 16384)

	for i := 0; i < 16384; i++ {
		indices[i] = randomBitString(TREE_DEPTH)
		data[i] = fmt.Sprintf("angela%d", i)
	}

	benchmarkInsertN(tree, indices, data, b)
}

func BenchmarkBatchInsert64(b *testing.B) {
	tree, _ := makeTree()

	transactions := make([]*transaction, 64)

	for i := 0; i < 64; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := transaction{id: index, data: d}
		transactions[i] = &t
	}

	b.ResetTimer()
	tree.batchInsert(transactions)
}

func BenchmarkBatchInsert128(b *testing.B) {
	tree, _ := makeTree()

	transactions := make([]*transaction, 128)

	for i := 0; i < 128; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := transaction{id: index, data: d}
		transactions[i] = &t
	}

	b.ResetTimer()
	tree.batchInsert(transactions)
}

func BenchmarkBatchInsert256(b *testing.B) {
	tree, _ := makeTree()

	transactions := make([]*transaction, 256)

	for i := 0; i < 256; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := transaction{id: index, data: d}
		transactions[i] = &t
	}

	b.ResetTimer()
	tree.batchInsert(transactions)
}

func BenchmarkBatchInsert512(b *testing.B) {
	tree, _ := makeTree()

	transactions := make([]*transaction, 512)

	for i := 0; i < 512; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := transaction{id: index, data: d}
		transactions[i] = &t
	}

	b.ResetTimer()
	tree.batchInsert(transactions)
}

func BenchmarkBatchInsert1024(b *testing.B) {
	tree, _ := makeTree()

	transactions := make([]*transaction, 1024)

	for i := 0; i < 1024; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := transaction{id: index, data: d}
		transactions[i] = &t
	}

	b.ResetTimer()
	tree.batchInsert(transactions)
}

func BenchmarkBatchInsert2048(b *testing.B) {
	tree, _ := makeTree()

	transactions := make([]*transaction, 2048)

	for i := 0; i < 2048; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := transaction{id: index, data: d}
		transactions[i] = &t
	}

	b.ResetTimer()
	tree.batchInsert(transactions)
}

func BenchmarkBatchInsert4096(b *testing.B) {
	tree, _ := makeTree()

	transactions := make([]*transaction, 4096)

	for i := 0; i < 4096; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := transaction{id: index, data: d}
		transactions[i] = &t
	}

	b.ResetTimer()
	tree.batchInsert(transactions)
}

func BenchmarkBatchInsert8192(b *testing.B) {
	tree, _ := makeTree()

	transactions := make([]*transaction, 8192)

	for i := 0; i < 8192; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := transaction{id: index, data: d}
		transactions[i] = &t
	}

	b.ResetTimer()
	tree.batchInsert(transactions)
}

func BenchmarkBatchInsert16384(b *testing.B) {
	tree, _ := makeTree()

	transactions := make([]*transaction, 16384)

	for i := 0; i < 16384; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := transaction{id: index, data: d}
		transactions[i] = &t
	}

	b.ResetTimer()
	tree.batchInsert(transactions)
}

func BenchmarkBatch2Insert64(b *testing.B) {
	tree, _ := makeTree()

	transactions := make([]*transaction, 64)

	for i := 0; i < 64; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := transaction{id: index, data: d}
		transactions[i] = &t
	}

	b.ResetTimer()
	tree.batch2Insert(transactions)
}

func BenchmarkBatch2Insert128(b *testing.B) {
	tree, _ := makeTree()

	transactions := make([]*transaction, 128)

	for i := 0; i < 128; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := transaction{id: index, data: d}
		transactions[i] = &t
	}

	b.ResetTimer()
	tree.batch2Insert(transactions)
}

func BenchmarkBatch2Insert256(b *testing.B) {
	tree, _ := makeTree()

	transactions := make([]*transaction, 256)

	for i := 0; i < 256; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := transaction{id: index, data: d}
		transactions[i] = &t
	}

	b.ResetTimer()
	tree.batch2Insert(transactions)
}

func BenchmarkBatch2Insert512(b *testing.B) {
	tree, _ := makeTree()

	transactions := make([]*transaction, 512)

	for i := 0; i < 512; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := transaction{id: index, data: d}
		transactions[i] = &t
	}

	b.ResetTimer()
	tree.batch2Insert(transactions)
}

func BenchmarkBatch2Insert1024(b *testing.B) {
	tree, _ := makeTree()

	transactions := make([]*transaction, 1024)

	for i := 0; i < 1024; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := transaction{id: index, data: d}
		transactions[i] = &t
	}

	b.ResetTimer()
	tree.batch2Insert(transactions)
}

func BenchmarkBatch2Insert2048(b *testing.B) {
	tree, _ := makeTree()

	transactions := make([]*transaction, 2048)

	for i := 0; i < 2048; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := transaction{id: index, data: d}
		transactions[i] = &t
	}

	b.ResetTimer()
	tree.batch2Insert(transactions)
}

func BenchmarkBatch2Insert4096(b *testing.B) {
	tree, _ := makeTree()

	transactions := make([]*transaction, 4096)

	for i := 0; i < 4096; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := transaction{id: index, data: d}
		transactions[i] = &t
	}

	b.ResetTimer()
	tree.batch2Insert(transactions)
}

func BenchmarkBatch2Insert8192(b *testing.B) {
	tree, _ := makeTree()

	transactions := make([]*transaction, 8192)

	for i := 0; i < 8192; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := transaction{id: index, data: d}
		transactions[i] = &t
	}

	b.ResetTimer()
	tree.batch2Insert(transactions)
}

func BenchmarkBatch2Insert16384(b *testing.B) {
	tree, _ := makeTree()

	transactions := make([]*transaction, 16384)

	for i := 0; i < 16384; i++ {
		index := randomBitString(TREE_DEPTH)
		d := fmt.Sprintf("angela%d", i)
		t := transaction{id: index, data: d}
		transactions[i] = &t
	}

	b.ResetTimer()
	tree.batch2Insert(transactions)
}
