package types

func CalculateQuorum(n int) int {
	f := (n-1)/3
	return n-f
}
