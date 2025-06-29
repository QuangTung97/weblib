package sliceutil

import (
	"cmp"
	"slices"
)

func Map[A, B any](input []A, mapFn func(e A) B) []B {
	result := make([]B, 0, len(input))
	for _, e := range input {
		y := mapFn(e)
		result = append(result, y)
	}
	return result
}

func GetMapKeys[K cmp.Ordered, V any](m map[K]V) []K {
	result := make([]K, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	slices.Sort(result)
	return result
}
