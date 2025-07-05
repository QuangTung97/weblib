package null

import (
	"bytes"
	"database/sql"
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

	if val.CanInt() {
		return val.Int(), nil
	}
	if val.CanUint() {
		return int64(val.Uint()), nil
	}
	if val.CanFloat() {
		return val.Float(), nil
	}
	if val.Kind() == reflect.String {
		return val.String(), nil
	}
	if val.Kind() == reflect.Bool {
		return val.Bool(), nil
	}

	// implement driver.Valuer interface
	dataInterface := any(n.Data)
	if valuer, ok := dataInterface.(driver.Valuer); ok {
		return valuer.Value()
	}

	// check is time type
	var emptyTime time.Time
	timeType := reflect.TypeOf(emptyTime)
	if val.CanConvert(timeType) {
		timeVal := val.Convert(timeType)
		return timeVal.Interface().(time.Time), nil
	}

	// check is byte slice
	var emptySlice []byte
	sliceType := reflect.TypeOf(emptySlice)
	if val.CanConvert(sliceType) {
		sliceVal := val.Convert(sliceType)
		return sliceVal.Interface().([]byte), nil
	}

	return nil, fmt.Errorf("unsupported sql value for null.Null[%s] type", val.Type().String())
}

func (n *Null[T]) Scan(inputValue interface{}) error {
	if inputValue == nil {
		n.Valid = false
		var empty T
		n.Data = empty
		return nil
	}

	var dataContent T
	dataVal := reflect.ValueOf(&dataContent).Elem()

	genericError := fmt.Errorf(
		"failed to scan type '%s' into null.Null[%s]",
		reflect.TypeOf(inputValue).String(),
		dataVal.Type().String(),
	)

	dataInterface := any(&dataContent)
	if scanner, ok := dataInterface.(sql.Scanner); ok {
		if err := scanner.Scan(inputValue); err != nil {
			return err
		}
		n.Valid = true
		n.Data = dataContent
		return nil
	}

	switch x := inputValue.(type) {
	case int64:
		if err := scanInt64ToData(dataVal, x, genericError); err != nil {
			return err
		}

	case string:
		if err := scanStringToData(dataVal, x, genericError); err != nil {
			return err
		}

	case time.Time:
		if err := scanTimeToData(dataVal, x, genericError); err != nil {
			return err
		}

	case []byte:
		if err := scanSliceToData(dataVal, x, genericError); err != nil {
			return err
		}

	case float64:
		if err := scanFloat64ToData(dataVal, x, genericError); err != nil {
			return err
		}

	case bool:
		if err := scanBoolToData(dataVal, x, genericError); err != nil {
			return err
		}

	default:
		return genericError
	}

	n.Valid = true
	n.Data = dataContent
	return nil
}

func scanInt64ToData(dataVal reflect.Value, x int64, genericErr error) error {
	lostPrecisionErr := fmt.Errorf("lost precision when scan null.Null[%s]", dataVal.Type().String())

	if dataVal.CanInt() {
		dataVal.SetInt(x)
		if dataVal.Int() != x {
			return lostPrecisionErr
		}
		return nil
	}

	if dataVal.CanUint() {
		dataVal.SetUint(uint64(x))
		if dataVal.Uint() != uint64(x) {
			return lostPrecisionErr
		}
		return nil

	}

	return genericErr
}

func scanStringToData(dataVal reflect.Value, x string, genericErr error) error {
	if dataVal.Kind() != reflect.String {
		return genericErr
	}

	dataVal.SetString(x)
	return nil
}

func scanTimeToData(dataVal reflect.Value, x time.Time, genericErr error) error {
	dataType := dataVal.Type()
	inputVal := reflect.ValueOf(x)

	if !inputVal.CanConvert(dataType) {
		return genericErr
	}

	dataVal.Set(inputVal.Convert(dataType))
	return nil
}

func scanSliceToData(dataVal reflect.Value, x []byte, genericErr error) error {
	dataType := dataVal.Type()
	inputVal := reflect.ValueOf(x)

	if !inputVal.CanConvert(dataType) {
		return genericErr
	}

	dataVal.Set(inputVal.Convert(dataType))
	return nil
}

func scanFloat64ToData(dataVal reflect.Value, x float64, genericErr error) error {
	if !dataVal.CanFloat() {
		return genericErr
	}
	dataVal.SetFloat(x)
	return nil
}

func scanBoolToData(dataVal reflect.Value, x bool, genericErr error) error {
	if dataVal.Kind() != reflect.Bool {
		return genericErr
	}
	dataVal.SetBool(x)
	return nil
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
