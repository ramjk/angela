package main

import (
	// "fmt"
	"encoding/json"
    "log"
    "net/http"
    "github.com/gorilla/mux"
)

type Transaction struct {
	MultiLeaf bool
	TransactionType string
	Index string
}

func Practice(w http.ResponseWriter, r *http.Request) {
	data := Transaction{MultiLeaf: true, TransactionType: "w", Index: "1001"}
    json.NewEncoder(w).Encode(data)
}


func main() {
	router := mux.NewRouter()
	router.HandleFunc("/merkletree", Practice).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}