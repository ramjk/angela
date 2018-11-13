package merkle

import "math/big"
import "fmt"

type ProofType bool

const (
	NONMEMBERSHIP    ProofType = false
	MEMBERSHIP       ProofType = true
)

type CoPathPair struct {
	ID        string
	digest    []byte
}

type Proof struct {
	proofType    ProofType
	queryID      string
	proofID      string
	coPath       []CoPathNode
}

func MakeProof(proofType ProofType, queryID string, proofID big.Int, coPath []CoPathNode) (*Proof) {
	return &Proof{proofType, queryID, proofID, coPath}
}
