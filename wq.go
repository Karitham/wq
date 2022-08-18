package wq

import (
	"runtime"
	"sync/atomic"
)

// WQueue is a queue of work to be done (with associated workers).
type WQueue[T any] struct {
	closed atomic.Bool
	q      []atomic.Pointer[T]
}

// New returns a new WorkQueue of the given worker count.
func New[T any](workerCount int) *WQueue[T] {
	return &WQueue[T]{
		closed: atomic.Bool{},
		q:      make([]atomic.Pointer[T], workerCount),
	}
}

// EnQ adds a work item to the queue.
func (wq *WQueue[T]) EnQ(item *T) {
	for i := 0; !wq.Closed(); i = (i + 1) % len(wq.q) {
		if wq.q[i].CompareAndSwap(nil, item) {
			return
		}
	}
}

// DeQ removes and returns a work item from the queue.
func (wq *WQueue[T]) DeQ(i int) (t *T, valid bool) {
	for !wq.Closed() {
		item := wq.q[i].Load()
		if item == nil {
			runtime.Gosched()
			continue
		}

		if wq.q[i].CompareAndSwap(item, nil) {
			return item, !wq.Closed()
		}
	}

	item := wq.q[i].Load()
	if item == nil {
		return nil, !wq.Closed()
	}

	return item, !wq.Closed()
}

// Close closes the queue.
func (wq *WQueue[T]) Close() {
	wq.closed.Store(true)
}

// Closed returns true if the queue is closed.
func (wq *WQueue[T]) Closed() bool {
	return wq.closed.Load()
}

// Drain removes all items from the queue and returns them.
func (wq *WQueue[T]) Drain() []T {
	var items []T
	for i := 0; i < len(wq.q); i++ {
		item := wq.q[i].Load()
		if item != nil {
			items = append(items, *item)
			wq.q[i].Store(nil)
		}
	}
	return items
}

// Run starts workers and executes the given work function for each item in the queue.
func (wq *WQueue[T]) Run(fn func(*T)) {
	for i := 0; i < len(wq.q); i++ {
		go func(i int) {
			for v, ok := wq.DeQ(i); ok; v, ok = wq.DeQ(i) {
				fn(v)
			}
		}(i)
	}
}
