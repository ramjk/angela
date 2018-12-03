package main

import (
	"../merkle"
	"fmt"
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
	data := Transaction{MultiLeaf: true, TransactionType: "W", Index: "practice"}
	fmt.Println(data)
    json.NewEncoder(w).Encode(data)
}

func GenerateProof(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if params["TransactionType"] != "R" {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		T, _ := merkle.MakeTree()
		fmt.Println(params["Index"])
		proof := T.GenerateProof(params["Index"])
		json.NewEncoder(w).Encode(proof)
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/merkletree", Practice).Methods("GET")
	router.HandleFunc("/merkletree/proof", GenerateProof).Methods("GET").Queries("MultiLeaf", "{MultiLeaf}", "TransactionType", "{TransactionType}", "Index", "{Index}")
	log.Fatal(http.ListenAndServe(":8000", router))
}