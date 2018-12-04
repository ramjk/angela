package main

import (
	"../merkle"
	"fmt"
	"encoding/json"
	"encoding/base64"
    "log"
    "net/http"
    "github.com/gorilla/mux"
)

var Angela *merkle.SparseMerkleTree = merkle.MakeTree()

type Transaction struct {
	MultiLeaf bool
	TransactionType string
	Index string
}

type WriteTransaction struct {
	MultiLeaf bool
	TransactionType string
	Index string
	Data string
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
		fmt.Println("Angela root []byte")
		fmt.Println(base64.StdEncoding.EncodeToString(Angela.GetRoot()))
		proof := Angela.GenerateProof(params["Index"])
		json.NewEncoder(w).Encode(proof)
	}
}

func UpdateLeaf(w http.ResponseWriter, r *http.Request) {
	var tx WriteTransaction
	_ = json.NewDecoder(r.Body).Decode(&tx)

	if tx.TransactionType != "W" {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		ok := Angela.Insert(tx.Index, tx.Data)
		if ok {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		
	}
}

func GetRoot(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Angela.GetRoot())
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/merkletree", Practice).Methods("GET")
	router.HandleFunc("/merkletree/root", GetRoot).Methods("GET")
	router.HandleFunc("/merkletree/prove", GenerateProof).Methods("GET").Queries("MultiLeaf", "{MultiLeaf}", "TransactionType", "{TransactionType}", "Index", "{Index}")
	router.HandleFunc("/merkletree/update", UpdateLeaf).Methods("POST")
	log.Fatal(http.ListenAndServe(":8000", router))
}