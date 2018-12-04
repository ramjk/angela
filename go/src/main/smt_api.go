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
func BatchWrite(transactionsKeys []*C.char, transactionsValues []*C.char) uintptr {
	tree, _ := merkle.MakeTree()

	transactions := make([]*merkle.Transaction, len(transactionsKeys))

	fmt.Println(len(transactionsKeys))

	for i:=0; i < len(transactionsKeys); i++ {
		transactions[i] = &merkle.Transaction{C.GoString(transactionsKeys[i]), C.GoString(transactionsValues[i])}
		fmt.Println(C.GoString(transactionsKeys[i]))
		fmt.Println(len(C.GoString(transactionsKeys[i])))
		fmt.Println(C.GoString(transactionsValues[i]))
	}
	worked, err := tree.BatchInsert(transactions)

	onesSlice := make([]bool, len(transactionsKeys))
	for i := range onesSlice {
    	onesSlice[i] = worked
	}
	if err != nil {
		return uintptr(unsafe.Pointer(&onesSlice[0]))
	}
	return uintptr(unsafe.Pointer(&onesSlice[0]))
}

//export Read
func Read(nodeId *C.char) **C.char {
	tree, _ := merkle.MakeTree()

	id := C.GoString(nodeId)
    results := tree.CGenerateProof(id)
    fmt.Println("ANEESH")
    fmt.Println(results)
    resultLength := len(results)

    // allocate space for all the node ids and digests and other fields in proof
    cArray := C.malloc(C.size_t(resultLength) * C.size_t(unsafe.Sizeof(uintptr(0))))
    fmt.Println(cArray)
    // convert the C array to a Go Array so we can index it
    a := (*[1<<30]*C.char)(cArray)
    // a[0] = C.CString(proof.proofType.String())
    // a[1] = C.CString(proof.queryID)
    // a[2] = C.CString(proof.proofID)

    // for i := 3; i < resultLength; i += 2 {
    // 	a[i] = C.CString(proof.coPath[i-3].ID)
    // 	a[i+1] = C.CString(proof.coPath[i-3].digest.String())
    // }
    for i := 0; i < resultLength; i++ {
    	a[i] = C.CString(results[i])
    }
    return (**C.char)(cArray)
}

//export FreeCPointers
func FreeCPointers(pointer **C.char, numItems int) {
	p := unsafe.Pointer(pointer)
    a := (*[1<<30]*C.char)(p)[:numItems:numItems]
    // fmt.Println(a)
    // fmt.Println(numItems)

    for idx:=0; idx<numItems; idx++  {
    	// fmt.Println(idx)
    	// fmt.Println(a[idx])
    	// fmt.Println(C.GoString(a[idx]))
    	fmt.Println("About to free")
    	fmt.Println(a[idx])
        C.free(unsafe.Pointer(a[idx]))
    }
    fmt.Println("About to free")
    fmt.Println(p)
    C.free(p)
}

func main(){}
