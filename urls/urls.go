package urls

import (
	"fmt"
	"iter"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/QuangTung97/weblib/null"
)

func New[T any](pattern string) *Path[T] {
	var empty T

	jsonTagSet := map[string]struct{}{}
	for jsonTag := range getAllJsonTags(empty) {
		if jsonTag.err != nil {
			panic(jsonTag.err.Error())
		}
		jsonTagSet[jsonTag.name] = struct{}{}
	}

	for param := range findPathParams(pattern) {
		_, ok := jsonTagSet[param.name]
		if !ok {
			structName := reflect.TypeOf(empty).Name()
			panic(fmt.Sprintf("missing json tag '%s' in struct '%s'", param.name, structName))
		}
	}

	return &Path[T]{
		pattern: pattern,
	}
}

type Path[T any] struct {
	pattern string
}

func (p Path[T]) Eval(params T) string {
	jsonTagMap := map[string]jsonTagValue{}
	for jsonTag := range getAllJsonTags(params) {
		jsonTagMap[jsonTag.name] = jsonTag
	}

	var buf strings.Builder
	lastIndex := 0
	for param := range findPathParams(p.pattern) {
		buf.WriteString(p.pattern[lastIndex:param.begin])

		jsonTag := jsonTagMap[param.name]
		delete(jsonTagMap, param.name)
		buf.WriteString(jsonTag.value)

		lastIndex = param.end
	}

	queryParams := url.Values{}
	for _, jsonTag := range jsonTagMap {
		if jsonTag.isZero {
			continue
		}
		queryParams.Add(jsonTag.name, jsonTag.value)
	}

	u := url.URL{
		Path:     buf.String(),
		RawQuery: queryParams.Encode(),
	}
	return u.String()
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
			err := fmt.Errorf("params type '%s' is not a struct", valType.Name())
			yield(jsonTagValue{err: err})
		}
	}

	return func(yield func(jsonTagValue) bool) {
		for index := range value.NumField() {
			fieldType := valType.Field(index)
			fieldVal := value.Field(index)

			jsonTag := fieldType.Tag.Get("json")
			if len(jsonTag) == 0 {
				err := fmt.Errorf(
					"missing json tag of field '%s' in struct '%s'",
					fieldType.Name, valType.Name(),
				)
				yield(jsonTagValue{err: err})
				return
			}

			valueStr, ok := reflectValueToString(fieldVal)
			if !ok {
				err := fmt.Errorf(
					"not support type '%s' of field '%s' in struct '%s'",
					fieldType.Type.String(), fieldType.Name, valType.Name(),
				)
				yield(jsonTagValue{err: err})
				return

			}

			tagVal := jsonTagValue{
				name:   jsonTag,
				value:  valueStr,
				isZero: fieldVal.IsZero(),
			}

			if !yield(tagVal) {
				return
			}
		}
	}
}

func reflectValueToString(val reflect.Value) (string, bool) {
	switch val.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(val.Bool()), true

	case reflect.String:
		return val.String(), true

	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10), true

	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(val.Uint(), 10), true

	default:
		output, ok := null.IsNullType(val)
		if !ok {
			return "", false
		}

		if output.NonNull {
			return reflectValueToString(output.DataField)
		}

		return "null", true
	}
}
