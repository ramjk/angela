package merkle

type transaction struct {
	id string
	data string
}

type batchedTransaction []*transaction

func (s batchedTransaction) Len() int {
    return len(s)
}
func (s batchedTransaction) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}
func (s batchedTransaction) Less(i, j int) bool {
    return s[i].id < s[j].id
}