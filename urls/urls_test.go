package urls

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/QuangTung97/weblib/null"
)

type testParams struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
	Val  string `json:"val"`
}

func TestPathNew(t *testing.T) {
	t.Run("not a struct", func(t *testing.T) {
		assert.PanicsWithValue(t, "params type 'int' is not a struct", func() {
			New[int]("/home")
		})
	})

	t.Run("missing tag name", func(t *testing.T) {
		type invalidStruct struct {
			ID int
		}
		assert.PanicsWithValue(t, "missing json tag of field 'ID' in struct 'invalidStruct'", func() {
			New[invalidStruct]("/home")
		})
	})

	t.Run("not support type", func(t *testing.T) {
		type invalidStruct struct {
			ID   int     `json:"id"`
			Name *string `json:"name"`
		}
		assert.PanicsWithValue(
			t,
			"not support type '*string' of field 'Name' in struct 'invalidStruct'",
			func() {
				New[invalidStruct]("/home")
			},
		)
	})

	t.Run("missing json tag in path param", func(t *testing.T) {
		assert.PanicsWithValue(
			t,
			"missing json tag 'pno' in struct 'testParams'",
			func() {
				New[testParams]("/products/{pno}")
			},
		)
	})
}

func TestPath_Eval(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		p := New[testParams]("/users/{id}")
		newURL := p.Eval(testParams{
			ID:   11,
			Name: "hello",
		})
		assert.Equal(t, "/users/11?name=hello", newURL)
	})

	t.Run("with special string", func(t *testing.T) {
		p := New[testParams]("/users%/{id}")
		newURL := p.Eval(testParams{
			ID:   11,
			Name: "hello$",
		})
		assert.Equal(t, "/users%25/11?name=hello%24", newURL)
	})

	t.Run("with empty struct", func(t *testing.T) {
		p := New[testParams]("/users/{id}/members/{val}")
		newURL := p.Eval(testParams{
			Val: "test",
		})
		assert.Equal(t, "/users/0/members/test", newURL)
		assert.Equal(t, "/users/{id}/members/{val}", p.GetPattern())
		assert.Equal(t, []string{"id", "val"}, p.GetPathParams())
		assert.Equal(t, []string{"name", "age"}, p.GetNonPathParams())
	})
}

func findPathParamsList(pattern string) []pathParam {
	return slices.Collect(findPathParams(pattern))
}

func TestFindPathParams(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		result := findPathParamsList("/users/{id}")
		assert.Equal(t, []pathParam{
			{
				name:  "id",
				begin: len("/users/"),
				end:   len("/users/{id}"),
			},
		}, result)
	})

	t.Run("multi params", func(t *testing.T) {
		result := findPathParamsList("/users/{id}/role/{val}")
		assert.Equal(t, []pathParam{
			{
				name:  "id",
				begin: len("/users/"),
				end:   len("/users/{id}"),
			},
			{
				name:  "val",
				begin: len("/users/{id}/role/"),
				end:   len("/users/{id}/role/{val}"),
			},
		}, result)
	})

	t.Run("multi params, with break", func(t *testing.T) {
		it := findPathParams("/users/{id}/role/{val}")
		var result []pathParam
		for param := range it {
			result = append(result, param)
			break
		}
		assert.Equal(t, []pathParam{
			{
				name:  "id",
				begin: len("/users/"),
				end:   len("/users/{id}"),
			},
		}, result)
	})

	t.Run("no param", func(t *testing.T) {
		result := findPathParamsList("/users/home")
		assert.Equal(t, []pathParam(nil), result)
	})

	t.Run("single param empty", func(t *testing.T) {
		result := findPathParamsList("/users/{}")
		assert.Equal(t, []pathParam{
			{
				begin: len("/users/"),
				end:   len("/users/{}"),
			},
		}, result)
	})
}

func getAllJsonTagsList(obj any) []jsonTagValue {
	result := slices.Collect(getAllJsonTags(obj))

	// clear fieldValue
	for i := range result {
		result[i].fieldValue = reflect.Value{}
	}

	return result
}

func TestGetAllJSONTags(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		type entity struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Age   int64  `json:"age"`
			Title string `json:"title_test"`
			Count uint32 `json:"count"`

			Desc     null.Null[string] `json:"desc"`
			MemberID null.Null[int32]  `json:"member_id"`

			Checked bool `json:"checked"`
		}

		// get all
		result := getAllJsonTagsList(entity{
			ID:    11,
			Name:  "hello",
			Age:   21,
			Count: 130,

			Desc:    null.New("sample desc"),
			Checked: true,
		})
		assert.Equal(t, []jsonTagValue{
			{name: "id", value: "11"},
			{name: "name", value: "hello"},
			{name: "age", value: "21"},
			{name: "title_test", value: "", isZero: true},
			{name: "count", value: "130"},
			{name: "desc", value: "sample desc"},
			{name: "member_id", value: "null", isZero: true},
			{name: "checked", value: "true"},
		}, result)

		// get with break
		result = nil
		for jsonTag := range getAllJsonTags(entity{}) {
			jsonTag.fieldValue = reflect.Value{}
			result = append(result, jsonTag)
			break
		}
		assert.Equal(t, []jsonTagValue{
			{name: "id", value: "0", isZero: true},
		}, result)
	})
}

func TestSetStructWithValues(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		var params testParams
		values := map[string]string{
			"id":   "123",
			"name": "hello01",
			"age":  "21",
		}

		err := SetStructWithValues(&params, []string{"id", "name"}, func(name string) string {
			return values[name]
		})
		assert.Equal(t, nil, err)
		assert.Equal(t, testParams{
			ID:   123,
			Name: "hello01",
		}, params)
	})

	t.Run("normal, empty value", func(t *testing.T) {
		var params testParams
		err := SetStructWithValues(&params, []string{"id", "name"}, func(name string) string {
			return ""
		})
		assert.Equal(t, nil, err)
		assert.Equal(t, testParams{}, params)
	})

	t.Run("obj not a pointer", func(t *testing.T) {
		var params testParams
		err := SetStructWithValues(params, []string{"id", "name"}, func(name string) string {
			return ""
		})
		assert.Equal(t, errors.New("input object type 'urls.testParams' must be a pointer instead"), err)
		assert.Equal(t, testParams{}, params)
	})

	t.Run("missing json tag", func(t *testing.T) {
		type invalidEntity struct {
			ID int
		}
		var params invalidEntity
		err := SetStructWithValues(&params, []string{"id", "name"}, func(name string) string {
			return ""
		})
		assert.Equal(t, errors.New("missing json tag of field 'ID' in struct 'invalidEntity'"), err)
		assert.Equal(t, invalidEntity{}, params)
	})

	t.Run("can not convert from string to int", func(t *testing.T) {
		var params testParams
		values := map[string]string{
			"id": "123a",
		}
		err := SetStructWithValues(&params, []string{"id", "name"}, func(name string) string {
			return values[name]
		})
		assert.Equal(t, errors.New("can not set value '123a' to field 'id' with type 'int'"), err)
		assert.Equal(t, testParams{}, params)
	})

	t.Run("with boolean", func(t *testing.T) {
		type Optional bool
		type testEntity struct {
			Checked Optional `json:"checked"`
		}

		var params testEntity
		values := map[string]string{
			"checked": "true",
		}
		// success
		err := SetStructWithValues(&params, []string{"checked"}, func(name string) string {
			return values[name]
		})
		assert.Equal(t, nil, err)
		assert.Equal(t, testEntity{
			Checked: true,
		}, params)

		// error
		err = SetStructWithValues(&params, []string{"checked"}, func(name string) string {
			return "xxx"
		})
		assert.Equal(t, errors.New("can not set value 'xxx' to field 'checked' with type 'urls.Optional'"), err)
	})

	t.Run("with uint64", func(t *testing.T) {
		type testEntity struct {
			Age uint64 `json:"age"`
		}

		var params testEntity
		err := SetStructWithValues(&params, []string{"age"}, func(name string) string {
			return "51"
		})
		assert.Equal(t, nil, err)
		assert.Equal(t, testEntity{Age: 51}, params)

		// error
		err = SetStructWithValues(&params, []string{"age"}, func(name string) string {
			return "xxx"
		})
		assert.Equal(t, errors.New("can not set value 'xxx' to field 'age' with type 'uint64'"), err)
	})

	t.Run("null int", func(t *testing.T) {
		type testEntity struct {
			Age null.Null[int] `json:"age"`
		}

		var params testEntity
		err := SetStructWithValues(&params, []string{"age"}, func(name string) string {
			return "51"
		})
		assert.Equal(t, nil, err)
		assert.Equal(t, testEntity{Age: null.New(51)}, params)

		// error
		err = SetStructWithValues(&params, []string{"age"}, func(name string) string {
			return "xxx"
		})
		assert.Equal(t, errors.New("can not set value 'xxx' to field 'age' with type 'null.Null[int]'"), err)

		// set empty
		params = testEntity{}
		err = SetStructWithValues(&params, []string{"age"}, func(name string) string {
			return ""
		})
		assert.Equal(t, nil, err)
		assert.Equal(t, testEntity{}, params)
	})

	t.Run("invalid type", func(t *testing.T) {
		type testEntity struct {
			Age *int `json:"age"`
		}

		// error
		var params testEntity
		err := SetStructWithValues(&params, []string{"age"}, func(name string) string {
			return "xxx"
		})
		assert.Equal(t, errors.New("not support type '*int' of field 'Age' in struct 'testEntity'"), err)
	})
}
