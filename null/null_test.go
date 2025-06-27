package null

import (
	"encoding/json"
	"reflect"
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

func TestIsNullType(t *testing.T) {
	t.Run("valid, non null", func(t *testing.T) {
		x := New("hello")

		output, ok := IsNullType(reflect.ValueOf(x))
		assert.Equal(t, true, ok)
		assert.Equal(t, true, output.NonNull)

		assert.Equal(t, reflect.String, output.DataField.Kind())
		assert.Equal(t, "hello", output.DataField.String())

		assert.Equal(t, reflect.Bool, output.ValidField.Kind())
		assert.Equal(t, true, output.ValidField.Bool())
	})

	t.Run("valid and null", func(t *testing.T) {
		var x Null[string]
		output, ok := IsNullType(reflect.ValueOf(x))
		assert.Equal(t, true, ok)
		assert.Equal(t, false, output.NonNull)

		assert.Equal(t, reflect.String, output.DataField.Kind())
		assert.Equal(t, "", output.DataField.String())

		assert.Equal(t, reflect.Bool, output.ValidField.Kind())
		assert.Equal(t, false, output.ValidField.Bool())
	})

	t.Run("not a struct", func(t *testing.T) {
		x := "invalid"
		_, ok := IsNullType(reflect.ValueOf(x))
		assert.Equal(t, false, ok)
	})

	t.Run("only has one field", func(t *testing.T) {
		type invalidStruct struct {
			Valid bool
		}
		var x invalidStruct
		_, ok := IsNullType(reflect.ValueOf(x))
		assert.Equal(t, false, ok)
	})

	t.Run("first field is not boolean", func(t *testing.T) {
		type invalidStruct struct {
			Valid string
			Data  string
		}
		var x invalidStruct
		_, ok := IsNullType(reflect.ValueOf(x))
		assert.Equal(t, false, ok)
	})

	t.Run("first field is not named 'Valid'", func(t *testing.T) {
		type invalidStruct struct {
			Valid1 bool
			Data   string
		}
		var x invalidStruct
		_, ok := IsNullType(reflect.ValueOf(x))
		assert.Equal(t, false, ok)
	})

	t.Run("second field is not named 'Data'", func(t *testing.T) {
		type invalidStruct struct {
			Valid bool
			Data1 string
		}
		var x invalidStruct
		_, ok := IsNullType(reflect.ValueOf(x))
		assert.Equal(t, false, ok)
	})
}
