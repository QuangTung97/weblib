package null

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
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

func (n Null[T]) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}

	val := reflect.ValueOf(n.Data)

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		return val.Int(), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int64(val.Uint()), nil

	case reflect.Float32, reflect.Float64:
		return val.Float(), nil

	case reflect.String:
		return val.String(), nil

	case reflect.Bool:
		return val.Bool(), nil

	default:
		valType := val.Type()

		// implement driver.Valuer interface
		dataInterface := any(n.Data)
		if valuer, ok := dataInterface.(driver.Valuer); ok {
			return valuer.Value()
		}

		// check is time type
		var emptyTime time.Time
		timeType := reflect.TypeOf(emptyTime)
		if valType.ConvertibleTo(timeType) {
			timeVal := val.Convert(timeType)
			return timeVal.Interface().(time.Time), nil
		}

		// check is byte slice
		var emptySlice []byte
		sliceType := reflect.TypeOf(emptySlice)
		if valType.ConvertibleTo(sliceType) {
			sliceVal := val.Convert(sliceType)
			return sliceVal.Interface().([]byte), nil
		}

		return nil, fmt.Errorf("unsupported sql value for null.Null[%s] type", val.Type().String())
	}
}

type CheckNullOutput struct {
	NonNull    bool
	ValidField reflect.Value
	DataField  reflect.Value
}

func IsNullType(val reflect.Value) (CheckNullOutput, bool) {
	if val.Kind() != reflect.Struct {
		return CheckNullOutput{}, false
	}

	valType := val.Type()
	if valType.NumField() != 2 {
		return CheckNullOutput{}, false
	}

	firstField := val.Field(0)
	firstFieldType := valType.Field(0)

	secondField := val.Field(1)
	secondFieldType := valType.Field(1)

	if firstField.Kind() != reflect.Bool {
		return CheckNullOutput{}, false
	}
	if firstFieldType.Name != "Valid" {
		return CheckNullOutput{}, false
	}
	if secondFieldType.Name != "Data" {
		return CheckNullOutput{}, false
	}

	output := CheckNullOutput{
		ValidField: firstField,
		DataField:  secondField,
	}

	if !firstField.Bool() {
		return output, true
	}

	output.NonNull = true
	return output, true
}
