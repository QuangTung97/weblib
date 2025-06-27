package null

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEqual(t *testing.T) {
	t.Run("both non null", func(t *testing.T) {
		a := New(10)
		b := New(10)
		assert.Equal(t, true, Equal(a, b))
	})

	t.Run("both null", func(t *testing.T) {
		var a, b Null[int]
		assert.Equal(t, true, Equal(a, b))
	})

	t.Run("one is not null", func(t *testing.T) {
		var a Null[int]
		b := New(10)
		assert.Equal(t, false, Equal(a, b))
		assert.Equal(t, false, Equal(b, a))
	})
}
