package main

import (
	"merkle"
	"C"
	"fmt"
)

//export BatchWrite
func BatchWrite(transactionsKeys []*C.char, transactionsValues []*C.char) bool {
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
	if err != nil {
		return false
	}

	return worked
}

func main(){}