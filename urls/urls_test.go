package urls

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/QuangTung97/weblib/null"
)

type testParams struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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
	return slices.Collect(getAllJsonTags(obj))
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
		}

		result := getAllJsonTagsList(entity{
			ID:    11,
			Name:  "hello",
			Age:   21,
			Count: 130,

			Desc: null.New("sample desc"),
		})
		assert.Equal(t, []jsonTagValue{
			{name: "id", value: "11"},
			{name: "name", value: "hello"},
			{name: "age", value: "21"},
			{name: "title_test", value: "", isZero: true},
			{name: "count", value: "130"},
			{name: "desc", value: "sample desc"},
			{name: "member_id", value: "null", isZero: true},
		}, result)
	})
}
