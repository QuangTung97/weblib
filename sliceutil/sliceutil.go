package sliceutil

func Map[A, B any](input []A, mapFn func(e A) B) []B {
	result := make([]B, 0, len(input))
	for _, e := range input {
		y := mapFn(e)
		result = append(result, y)
	}
	return result
}
