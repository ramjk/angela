package merkle

type ProofType bool

const (
	NONMEMBERSHIP    ProofType = false
	MEMBERSHIP       ProofType = true
)

type CoPathPair struct {
	ID        string
	Digest    []byte
}

type Proof struct {
	ProofType    ProofType
	QueryID      string
	ProofID      string
	CoPath       []CoPathPair
}

func MakeProof(proofType ProofType, queryID string, proofID string, coPath []CoPathPair) (*Proof) {
	return &Proof{proofType, queryID, proofID, coPath}
}
