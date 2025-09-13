package linked

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

type listEntry struct {
	val  int
	list ListHead
}

func newEntry(x int) *listEntry {
	return &listEntry{
		val: x,
	}
}

func listEntryFromList(n *ListHead) *listEntry {
	ptr := unsafe.Pointer(n)

	var empty listEntry
	offset := unsafe.Offsetof(empty.list)

	return (*listEntry)(unsafe.Pointer(uintptr(ptr) - offset))
}

func TestListEntryPtr(t *testing.T) {
	x := &listEntry{
		val: 10,
	}
	assert.Same(t, x, listEntryFromList(&x.list))
}

func TestListHead(t *testing.T) {
	var head ListHead
	head.Init()
	assert.Equal(t, true, head.IsEmpty())

	e1 := newEntry(11)
	e2 := newEntry(12)
	e3 := newEntry(13)
	e4 := newEntry(21)

	head.PushBack(&e1.list)
	head.PushBack(&e2.list)
	head.PushBack(&e3.list)
	head.PushFront(&e4.list)

	var result []int
	for n := range head.All() {
		entry := listEntryFromList(n)
		result = append(result, entry.val)
	}
	assert.Equal(t, []int{21, 11, 12, 13}, result)

	// with break
	result = nil
	for n := range head.All() {
		entry := listEntryFromList(n)
		result = append(result, entry.val)
		break
	}
	assert.Equal(t, []int{21}, result)

	e2.list.Remove()

	// get all again
	result = nil
	for n := range head.All() {
		entry := listEntryFromList(n)
		result = append(result, entry.val)
	}
	assert.Equal(t, []int{21, 11, 13}, result)

	// pop front
	x := head.PopFront()
	assert.Equal(t, 21, listEntryFromList(x).val)

	// get all again
	result = nil
	for n := range head.All() {
		entry := listEntryFromList(n)
		result = append(result, entry.val)
	}
	assert.Equal(t, []int{11, 13}, result)

	assert.Equal(t, false, head.IsEmpty())
}
