package condvar

import (
	"context"
	"sync"
	"unsafe"

	"github.com/QuangTung97/weblib/utils/linked"
)

type Cond struct {
	locker sync.Locker

	mut      sync.Mutex
	waitList linked.ListHead
}

func NewCond(l sync.Locker) *Cond {
	c := &Cond{
		locker: l,
	}
	c.waitList.Init()
	return c
}

func (c *Cond) Wait(ctx context.Context) error {
	entry := c.beginWaiting()
	c.locker.Unlock()

	select {
	case <-entry.waitCh:
		c.locker.Lock()
		return nil

	case <-ctx.Done():
		c.locker.Lock()
		c.handleCtxCancel(entry)
		return ctx.Err()
	}
}

func (c *Cond) Signal() {
	c.mut.Lock()
	c.notifySingleWithoutLock()
	c.mut.Unlock()
}

func (c *Cond) Broadcast() {
	c.mut.Lock()
	for !c.waitList.IsEmpty() {
		c.notifySingleWithoutLock()
	}
	c.mut.Unlock()
}

func (c *Cond) notifySingleWithoutLock() {
	if c.waitList.IsEmpty() {
		return
	}

	n := c.waitList.PopFront()
	entry := waitEntryFromList(n)
	entry.isRemoved = true

	close(entry.waitCh)
}

func (c *Cond) beginWaiting() *waitEntry {
	entry := &waitEntry{
		waitCh: make(chan struct{}),
	}

	c.mut.Lock()
	c.waitList.PushBack(&entry.waitList)
	c.mut.Unlock()

	return entry
}

func (c *Cond) handleCtxCancel(entry *waitEntry) {
	c.mut.Lock()
	defer c.mut.Unlock()

	if !entry.isRemoved {
		entry.isRemoved = true
		entry.waitList.Remove()
		return
	}

	c.notifySingleWithoutLock()
}

type waitEntry struct {
	waitCh    chan struct{}
	waitList  linked.ListHead
	isRemoved bool
}

func waitEntryFromList(n *linked.ListHead) *waitEntry {
	ptr := unsafe.Pointer(n)

	var empty waitEntry
	offset := unsafe.Offsetof(empty.waitList)

	return (*waitEntry)(unsafe.Pointer(uintptr(ptr) - offset))
}
