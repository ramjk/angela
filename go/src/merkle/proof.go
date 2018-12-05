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
	ProofLength  int
	CoPath       []CoPathPair
}

func MakeProof(proofType ProofType, queryID string, proofID string, proofLength int, coPath []CoPathPair) (*Proof) {
	return &Proof{proofType, queryID, proofID, proofLength, coPath}
}
