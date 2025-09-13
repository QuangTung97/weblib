package linked

import "iter"

type ListHead struct {
	next *ListHead
	prev *ListHead
}

func (h *ListHead) Init() {
	h.next = h
	h.prev = h
}

func (h *ListHead) PushBack(n *ListHead) {
	last := h.prev

	last.next = n
	n.next = h

	h.prev = n
	n.prev = last
}

func (h *ListHead) PushFront(n *ListHead) {
	first := h.next

	h.next = n
	n.next = first

	first.prev = n
	n.prev = h
}

func (h *ListHead) PopFront() *ListHead {
	front := h.next
	next := front.next

	h.next = next
	next.prev = h

	front.next = nil
	front.prev = nil

	return front
}

func (h *ListHead) IsEmpty() bool {
	return h.next == h
}

func (h *ListHead) All() iter.Seq[*ListHead] {
	return func(yield func(*ListHead) bool) {
		for n := h.next; n != h; n = n.next {
			if !yield(n) {
				return
			}
		}
	}
}

func (h *ListHead) Remove() {
	next := h.next
	prev := h.prev

	next.prev = prev
	prev.next = next

	h.next = nil
	h.prev = nil
}
