package types

func CalculateFault(n int) int {
	return (n-1)/3
}

func CalculateQuorum(n int) int {
	return n-CalculateFault(n)
}

func CalculateOneCorrect(n int) int {
	return CalculateFault(n)+1
}
