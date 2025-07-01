package null

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

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

type customValue struct {
	data int64
}

func (v customValue) Value() (driver.Value, error) {
	return v.data, nil
}

func TestNull_Value(t *testing.T) {
	t.Run("int64", func(t *testing.T) {
		// normal
		val, err := New[int64](1234).Value()
		assert.Equal(t, nil, err)
		assert.Equal(t, int64(1234), val)

		// null
		val, err = Null[int64]{}.Value()
		assert.Equal(t, nil, err)
		assert.Equal(t, nil, val)
	})

	t.Run("uint32", func(t *testing.T) {
		// normal
		val, err := New[uint32](1234).Value()
		assert.Equal(t, nil, err)
		assert.Equal(t, int64(1234), val)
	})

	t.Run("float32", func(t *testing.T) {
		// normal
		val, err := New[float32](1234).Value()
		assert.Equal(t, nil, err)
		assert.Equal(t, float64(1234), val)
	})

	t.Run("string", func(t *testing.T) {
		// normal
		val, err := New("abcd").Value()
		assert.Equal(t, nil, err)
		assert.Equal(t, "abcd", val)
	})

	t.Run("bool", func(t *testing.T) {
		// normal
		val, err := New(true).Value()
		assert.Equal(t, nil, err)
		assert.Equal(t, true, val)
	})

	t.Run("time", func(t *testing.T) {
		now := time.Now()
		type customType time.Time

		val, err := New[customType](customType(now)).Value()
		assert.Equal(t, nil, err)
		assert.Equal(t, now, val)
	})

	t.Run("byte slice", func(t *testing.T) {
		type customType []byte
		data := customType("hello")

		val, err := New(data).Value()
		assert.Equal(t, nil, err)
		assert.Equal(t, []byte("hello"), val)
	})

	t.Run("fixed array", func(t *testing.T) {
		data := [10]byte{11, 12, 13, 14, 15}

		val, err := New(data).Value()
		assert.Equal(t, errors.New("unsupported sql value for null.Null[[10]uint8] type"), err)
		assert.Equal(t, nil, val)
	})

	t.Run("custom type with driver.Value", func(t *testing.T) {
		x := customValue{data: 1234}
		val, err := New(x).Value()
		assert.Equal(t, nil, err)
		assert.Equal(t, int64(1234), val)
	})

	t.Run("null inside null", func(t *testing.T) {
		x := New(int64(21))
		val, err := New(x).Value()
		assert.Equal(t, nil, err)
		assert.Equal(t, int64(21), val)
	})
}
