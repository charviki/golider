package gtool

import "golang.org/x/exp/constraints"

func Ptr[T any](v T) *T {
	return &v
}

func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}
