package null

// Null for representing nullable value, is a better replacement for pointer
type Null[T any] struct {
	Valid bool
	Data  T
}

// New creates a non-null value
func New[T any](data T) Null[T] {
	return Null[T]{
		Valid: true,
		Data:  data,
	}
}

func Equal[T comparable](a, b Null[T]) bool {
	if !a.Valid && !b.Valid {
		// null = null
		return true
	}

	if !a.Valid {
		// null != non-null
		return false
	}

	if !b.Valid {
		// non-null != null
		return false
	}

	return a.Data == b.Data
}
