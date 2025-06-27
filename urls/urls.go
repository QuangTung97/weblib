package urls

import (
	"errors"
	"fmt"
	"iter"
	"reflect"
	"strconv"
)

func New[T any](pattern string) *Path[T] {
	return &Path[T]{}
}

type Path[T any] struct {
}

func (p Path[T]) Eval(params T) string {
	return ""
}

type pathParam struct {
	name  string
	begin int
	end   int
}

func findPathParams(pattern string) iter.Seq[pathParam] {
	return func(yield func(pathParam) bool) {
		strLen := len(pattern)

		findOpening := true
		beginIndex := -1

		for i := range strLen {
			if findOpening {
				if pattern[i] != '{' {
					continue
				}

				findOpening = false
				beginIndex = i
			} else {
				if pattern[i] != '}' {
					continue
				}

				findOpening = true
				param := pathParam{
					name:  pattern[beginIndex+1 : i],
					begin: beginIndex,
					end:   i + 1,
				}

				if !yield(param) {
					return
				}
			}
		}
	}
}

type jsonTagValue struct {
	name   string
	value  string
	isZero bool
	err    error
}

func getAllJsonTags(obj any) iter.Seq[jsonTagValue] {
	value := reflect.ValueOf(obj)
	valType := value.Type()

	if valType.Kind() != reflect.Struct {
		return func(yield func(jsonTagValue) bool) {
			yield(jsonTagValue{err: errors.New("params type is not a struct")})
		}
	}

	return func(yield func(jsonTagValue) bool) {
		for index := range value.NumField() {
			fieldType := valType.Field(index)
			fieldVal := value.Field(index)

			jsonTag := fieldType.Tag.Get("json")

			tagVal := jsonTagValue{
				name:   jsonTag,
				value:  reflectValueToString(fieldVal),
				isZero: fieldVal.IsZero(),
			}

			if !yield(tagVal) {
				return
			}
		}
	}
}

func reflectValueToString(val reflect.Value) string {
	switch val.Kind() {
	case reflect.String:
		return val.String()

	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10)

	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(val.Uint(), 10)

	default:
		return fmt.Sprintf("%v", val.Interface())
	}
}
