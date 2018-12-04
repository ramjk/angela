package merkle

type Transaction struct {
	Id string
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
    return s[i].Id < s[j].Id
}

func min(i int, j int) int {
	if i < j { return i }
	return  j
}