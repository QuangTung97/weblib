package null

import (
	"bytes"
	"encoding/json"
)

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

var nullBytes = []byte("null")

func (n Null[T]) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return nullBytes, nil
	}
	return json.Marshal(n.Data)
}

func (n *Null[T]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		*n = Null[T]{}
		return nil
	}

	if err := json.Unmarshal(data, &n.Data); err != nil {
		return err
	}

	n.Valid = true
	return nil
}
