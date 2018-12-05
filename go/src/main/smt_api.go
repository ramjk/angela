package main

// #include <stdio.h>
// #include <stdlib.h>
import "C"

import (
	"merkle"
	"fmt"
	"unsafe"
)

//export BatchWrite
func BatchWrite(prefix *C.char, transactionsKeys []*C.char, transactionsValues []*C.char, epochNumber uint64) *C.char {
	tree := merkle.MakeTree(C.GoString(prefix))

	transactions := make([]*merkle.Transaction, len(transactionsKeys))

	for i:=0; i < len(transactionsKeys); i++ {
		transactions[i] = &merkle.Transaction{C.GoString(transactionsKeys[i]), C.GoString(transactionsValues[i])}
	}

	root, _ := tree.BatchInsert(transactions, epochNumber)

	// onesSlice := make([]bool, len(transactionsKeys))
	// for i := range onesSlice {
 //    	onesSlice[i] = worked
	// }
	// if err != nil {
	// 	return uintptr(unsafe.Pointer(&onesSlice[0]))
	// }
	// return uintptr(unsafe.Pointer(&onesSlice[0]))
    return C.CString(root)
}

//export Read
func Read(nodeId *C.char) **C.char {
	tree := merkle.MakeTree("")

	id := C.GoString(nodeId)
    results := tree.CGenerateProof(id)
    resultLength := len(results)
    // fmt.Println("received", resultLength)

    // allocate space for all the node ids and digests and other fields in proof
    cArray := C.malloc(C.size_t(resultLength) * C.size_t(unsafe.Sizeof(uintptr(0))))
    //fmt.Println(cArray)
    // convert the C array to a Go Array so we can index it
    a := (*[1<<30]*C.char)(cArray)
    for i := 0; i < resultLength; i++ {
    	a[i] = C.CString(results[i])
    }
    return (**C.char)(cArray)
}

//export GetLatestRoot
func GetLatestRoot() *C.char {
    tree := merkle.MakeTree("")
    return C.CString(tree.GetLatestRoot())
}

//export FreeCPointers
func FreeCPointers(pointer **C.char, numItems int) {
	p := unsafe.Pointer(pointer)
    a := (*[1<<30]*C.char)(p)[:numItems:numItems]
    fmt.Println("About to free this many pointers", numItems+1)

    for idx:=0; idx<numItems; idx++  {
        C.free(unsafe.Pointer(a[idx]))
    }
    C.free(p)
}

func main(){}
