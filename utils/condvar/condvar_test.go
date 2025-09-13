package condvar

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCond_BeginWaiting(t *testing.T) {
	c := NewCond(nil)

	entry1 := c.beginWaiting()
	assert.Equal(t, false, entry1.isRemoved)
	assert.Equal(t, false, c.waitList.IsEmpty())

	c.Signal()

	assert.Equal(t, true, entry1.isRemoved)
	assert.Equal(t, true, c.waitList.IsEmpty())
}

func TestCond_Signal__Multi_Waiting(t *testing.T) {
	c := NewCond(nil)

	entry1 := c.beginWaiting()
	entry2 := c.beginWaiting()

	c.Signal()

	assert.Equal(t, true, entry1.isRemoved)
	assert.Equal(t, false, entry2.isRemoved)
	assert.Equal(t, false, c.waitList.IsEmpty())

	c.Signal()

	assert.Equal(t, true, entry1.isRemoved)
	assert.Equal(t, true, entry2.isRemoved)
	assert.Equal(t, true, c.waitList.IsEmpty())

	c.Signal()
	assert.Equal(t, true, c.waitList.IsEmpty())
}

func TestCond_Broadcast(t *testing.T) {
	c := NewCond(nil)

	entry1 := c.beginWaiting()
	entry2 := c.beginWaiting()

	c.Broadcast()

	assert.Equal(t, true, entry1.isRemoved)
	assert.Equal(t, true, entry2.isRemoved)
	assert.Equal(t, true, c.waitList.IsEmpty())
}

func TestCond_HandleCtxCancel__Single(t *testing.T) {
	c := NewCond(nil)

	entry1 := c.beginWaiting()
	c.handleCtxCancel(entry1)

	assert.Equal(t, true, entry1.isRemoved)
	assert.Equal(t, true, c.waitList.IsEmpty())
}

func TestCond_HandleCtxCancel__With_Other_Waiter(t *testing.T) {
	c := NewCond(nil)

	entry1 := c.beginWaiting()
	entry2 := c.beginWaiting()
	entry3 := c.beginWaiting()

	c.handleCtxCancel(entry1)

	assert.Equal(t, true, entry1.isRemoved)
	assert.Equal(t, false, entry2.isRemoved)
	assert.Equal(t, false, entry3.isRemoved)
	assert.Equal(t, false, c.waitList.IsEmpty())
}

func TestCond_HandleCtxCancel__After_Signal__With_Other_Waiter(t *testing.T) {
	c := NewCond(nil)

	entry1 := c.beginWaiting()
	entry2 := c.beginWaiting()
	entry3 := c.beginWaiting()

	c.Signal()
	c.handleCtxCancel(entry1)

	assert.Equal(t, true, entry1.isRemoved)
	assert.Equal(t, true, entry2.isRemoved)
	assert.Equal(t, false, entry3.isRemoved)
	assert.Equal(t, false, c.waitList.IsEmpty())
}

func TestCond_Wait(t *testing.T) {
	var finished bool
	var mut sync.Mutex
	cond := NewCond(&mut)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		mut.Lock()
		defer mut.Unlock()

		for !finished {
			if err := cond.Wait(context.Background()); err != nil {
				return
			}
		}
	}()

	time.Sleep(10 * time.Millisecond)

	mut.Lock()
	finished = true
	mut.Unlock()
	cond.Signal()

	wg.Wait()
}

func TestCond_Wait__Context_Cancel(t *testing.T) {
	var finished bool
	var mut sync.Mutex
	cond := NewCond(&mut)
	assert.Equal(t, true, cond.waitList.IsEmpty())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		mut.Lock()
		defer mut.Unlock()

		for !finished {
			if err := cond.Wait(ctx); err != nil {
				return
			}
		}
	}()

	wg.Wait()

	assert.Equal(t, true, cond.waitList.IsEmpty())
}
