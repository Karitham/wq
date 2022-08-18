package wq

import (
	"runtime"
	"sync/atomic"
)

// WQueue is a queue of work to be done (with associated workers).
type WQueue[T any] struct {
	q      []atomic.Pointer[T]
	closed atomic.Bool
	i      atomic.Uint32
	doneWs atomic.Uint32
}

type Option struct {
	// The number of workers to start.
	wCount int

	// Do not start any workers.
	noStart bool
}

// New returns a new WorkQueue of the given worker count.
func New[T any](w func(*T), opts ...func(*Option)) *WQueue[T] {
	o := &Option{wCount: runtime.NumCPU(), noStart: false}
	for _, opt := range opts {
		opt(o)
	}

	wq := &WQueue[T]{
		closed: atomic.Bool{},
		i:      atomic.Uint32{},
		q:      make([]atomic.Pointer[T], o.wCount),
	}

	if !o.noStart {
		wq.Run(w)
	}
	return wq
}

// WithWorkerCount sets the number of workers to start.
func WithWorkerCount(wCount int) func(*Option) {
	return func(o *Option) {
		o.wCount = wCount
	}
}

// WithNoStart disables starting any workers.
func WithNoStart() func(*Option) {
	return func(o *Option) {
		o.noStart = true
	}
}

// EnQ adds a work item to the queue.
func (wq *WQueue[T]) EnQ(item *T) {
	for !wq.Closed() {
		i := wq.i.Add(1) % uint32(len(wq.q))
		if wq.q[i].CompareAndSwap(nil, item) {
			return
		}
	}

	panic("enqueue on closed queue")
}

// deQ removes and returns a work item from the queue.
func (wq *WQueue[T]) deQ(i int) (t *T, valid bool) {
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

	return wq.q[i].Swap(nil), !wq.Closed()
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
	for i := range wq.q {
		go func(i int) {
			defer wq.doneWs.Add(1)
			for v, ok := wq.deQ(i); ; v, ok = wq.deQ(i) {
				if v == nil && !ok {
					return
				}

				fn(v)
			}
		}(i)
	}
}

// Wait waits for all workers to finish.
func (wq *WQueue[T]) Wait() {
	wq.Close()
	for wq.doneWs.Load() != uint32(len(wq.q)) {
	}
}
