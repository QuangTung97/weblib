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

func Filter[T any](input []T, pred func(x T) bool) []T {
	result := make([]T, 0, len(input))
	for _, x := range input {
		if pred(x) {
			result = append(result, x)
		}
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

func Unique[T comparable](list []T) []T {
	var result []T
	set := map[T]struct{}{}
	for _, x := range list {
		_, existed := set[x]
		if existed {
			continue
		}
		result = append(result, x)
		set[x] = struct{}{}
	}
	return result
}

func SliceToMap[T any, K comparable](list []T, getKey func(x T) K) map[K]T {
	result := make(map[K]T, len(list))
	for _, x := range list {
		key := getKey(x)
		result[key] = x
	}
	return result
}

func SliceToSet[T any, K comparable](list []T, getKey func(x T) K) map[K]struct{} {
	result := make(map[K]struct{}, len(list))
	for _, x := range list {
		key := getKey(x)
		result[key] = struct{}{}
	}
	return result
}

func SliceToMapList[T any, K comparable](list []T, getKey func(x T) K) map[K][]T {
	result := map[K][]T{}
	for _, x := range list {
		key := getKey(x)
		result[key] = append(result[key], x)
	}
	return result
}
