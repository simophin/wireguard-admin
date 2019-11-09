package utils

func ReverseLessFunc(less func(i, j int) bool) func(i, j int) bool {
	return func(i, j int) bool {
		return less(j, i)
	}
}
