package merkle

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
	coPath       []CoPathPair
}

func MakeProof(proofType ProofType, queryID string, proofID string, coPath []CoPathPair) (*Proof) {
	return &Proof{proofType, queryID, proofID, coPath}
}
