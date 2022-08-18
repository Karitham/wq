package wq

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestDrain(t *testing.T) {
	workSize := 100
	q := New(func(*int) { time.Sleep(time.Nanosecond) })

	for i := 0; i < workSize; i++ {
		i := i
		q.EnQ(&i)
		q.Drain()
	}
	q.Wait()
}

func BenchmarkQueue(b *testing.B) {
	q := New(func(*int) { time.Sleep(time.Nanosecond) })

	vals := make([]int, 0, b.N)
	for i := 0; i < b.N; i++ {
		vals = append(vals, i)
	}

	b.ResetTimer()
	for i := range vals {
		q.EnQ(&vals[i])
	}
	q.Wait()
}

func BenchmarkChans(b *testing.B) {
	workerCount := runtime.NumCPU()
	c := make(chan int, workerCount)
	w := func() {
		for _, ok := <-c; ok; _, ok = <-c {
			time.Sleep(time.Nanosecond)
		}
	}

	vals := make([]int, 0, b.N)
	for i := 0; i < b.N; i++ {
		vals = append(vals, i)
	}

	for i := 0; i < workerCount; i++ {
		go w()
	}

	b.ResetTimer()
	for i := range vals {
		c <- vals[i]
	}
}

func TestV(t *testing.T) {
	const n = 150
	c := uint32(0)
	q := New(func(v *int) { atomic.AddUint32(&c, 1); fmt.Println(*v) })

	for i := 0; i < n; i++ {
		i := i
		q.EnQ(&i)
	}
	q.Wait()

	if c != n {
		t.Errorf("expected %d, got %d", n, c)
	}
}
