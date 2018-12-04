package merkle

type Transaction struct {
	ID string
	Data string
}

type BatchedTransaction []*Transaction

func (s BatchedTransaction) Len() int {
    return len(s)
}
func (s BatchedTransaction) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}
func (s BatchedTransaction) Less(i, j int) bool {
    return s[i].ID < s[j].ID
}

func min(i int, j int) int {
	if i < j { return i }
	return  j
}