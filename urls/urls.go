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

func (p Path[T]) GetPathParams() []string {
	var result []string
	for param := range findPathParams(p.pattern) {
		result = append(result, param.name)
	}
	return result
}

func (p Path[T]) GetNonPathParams() []string {
	pathParamSet := map[string]struct{}{}
	for param := range findPathParams(p.pattern) {
		pathParamSet[param.name] = struct{}{}
	}

	var result []string
	var emptyParams T
	for jsonTag := range getAllJsonTags(emptyParams) {
		_, existed := pathParamSet[jsonTag.name]
		if existed {
			continue
		}
		result = append(result, jsonTag.name)
	}
	return result
}

func SetStructWithValues(
	obj any, updateFields []string,
	valueFunc func(name string) string,
) error {
	updateSet := map[string]struct{}{}
	for _, jsonTag := range updateFields {
		updateSet[jsonTag] = struct{}{}
	}

	objValuePtr := reflect.ValueOf(obj)
	if objValuePtr.Kind() != reflect.Ptr {
		return fmt.Errorf("input object type '%s' must be a pointer instead", objValuePtr.Type().String())
	}

	objValue := objValuePtr.Elem()

	for jsonTag := range getAllJsonTagsOfValue(objValue) {
		if jsonTag.err != nil {
			return jsonTag.err
		}

		_, ok := updateSet[jsonTag.name]
		if !ok {
			continue
		}

		valueStr := valueFunc(jsonTag.name)
		if len(valueStr) == 0 {
			continue
		}

		if ok := updateValueFromString(jsonTag.fieldValue, valueStr); !ok {
			return fmt.Errorf(
				"can not set value '%s' to field '%s' with type '%s'",
				valueStr, jsonTag.name, jsonTag.fieldValue.Type().String(),
			)
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
// Internal Implementation
// ---------------------------------------------------------------------------

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

	fieldValue reflect.Value
}

func getAllJsonTags(obj any) iter.Seq[jsonTagValue] {
	return getAllJsonTagsOfValue(reflect.ValueOf(obj))
}

func getAllJsonTagsOfValue(value reflect.Value) iter.Seq[jsonTagValue] {
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

				fieldValue: fieldVal,
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

func updateValueFromString(val reflect.Value, str string) bool {
	switch val.Kind() {
	case reflect.Bool:
		b, err := strconv.ParseBool(str)
		if err != nil {
			return false
		}
		val.SetBool(b)
		return true

	case reflect.String:
		val.SetString(str)
		return true

	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		num, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return false
		}
		val.SetInt(num)
		return true

	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		num, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return false
		}
		val.SetUint(num)
		return true

	default:
		output, ok := null.IsNullType(val)
		if !ok {
			return false
		}

		if ok := updateValueFromString(output.DataField, str); !ok {
			return false
		}
		output.ValidField.SetBool(true)
		return true
	}
}
