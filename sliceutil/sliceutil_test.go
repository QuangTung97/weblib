package sliceutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	x := Map([]int{11, 12, 13}, func(e int) string {
		return fmt.Sprintf("%03d", e)
	})
	assert.Equal(t, []string{"011", "012", "013"}, x)
}

func TestFilter(t *testing.T) {
	x := Filter([]int{11, 12, 13}, func(x int) bool {
		return x <= 12
	})
	assert.Equal(t, []int{11, 12}, x)
}

func TestGetMapKeys(t *testing.T) {
	x := map[string]int{
		"A": 11,
		"B": 12,
		"C": 13,
	}
	assert.Equal(t, []string{"A", "B", "C"}, GetMapKeys(x))
}

func TestUnique(t *testing.T) {
	list := Unique([]int{11, 12, 11, 13, 14, 11, 14})
	assert.Equal(t, []int{11, 12, 13, 14}, list)
}

type testUser struct {
	id   int
	name string
}

func TestSliceToMap(t *testing.T) {
	user1 := testUser{id: 11, name: "user01"}
	user2 := testUser{id: 12, name: "user02"}
	user3 := testUser{id: 13, name: "user03"}

	m := SliceToMap([]testUser{user1, user2, user3}, func(x testUser) int {
		return x.id
	})
	assert.Equal(t, map[int]testUser{
		11: user1,
		12: user2,
		13: user3,
	}, m)

	set := SliceToSet([]testUser{user1, user2, user3}, func(x testUser) int {
		return x.id
	})
	assert.Equal(t, map[int]struct{}{
		11: {},
		12: {},
		13: {},
	}, set)
}

func TestSliceToMapList(t *testing.T) {
	user1 := testUser{id: 11, name: "user01"}
	user2 := testUser{id: 12, name: "user02"}
	user3 := testUser{id: 13, name: "user03"}
	user4 := testUser{id: 12, name: "user04"}

	m := SliceToMapList([]testUser{user1, user2, user3, user4}, func(x testUser) int {
		return x.id
	})

	assert.Equal(t, map[int][]testUser{
		11: {user1},
		12: {user2, user4},
		13: {user3},
	}, m)
}
