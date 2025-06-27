package null

import (
	"encoding/json"
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

func TestNullJSON(t *testing.T) {
	t.Run("marshal and unmarshal null", func(t *testing.T) {
		var x Null[int]
		data, err := json.Marshal(x)
		assert.Equal(t, nil, err)
		assert.Equal(t, "null", string(data))

		newVal := New(123)
		err = json.Unmarshal(data, &newVal)
		assert.Equal(t, nil, err)
		assert.Equal(t, x, newVal)
	})

	t.Run("marshal and unmarshal non null", func(t *testing.T) {
		x := New("hello world")
		data, err := json.Marshal(x)
		assert.Equal(t, nil, err)
		assert.Equal(t, `"hello world"`, string(data))

		var newVal Null[string]
		err = json.Unmarshal(data, &newVal)
		assert.Equal(t, nil, err)
		assert.Equal(t, x, newVal)
	})
}
