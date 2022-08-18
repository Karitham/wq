package wq

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func TestQueue(t *testing.T) {
	workSize := 10000000

	q := New(func(*int) { time.Sleep(time.Nanosecond) })

	for i := 0; i < workSize; i++ {
		i := i
		q.EnQ(&i)
	}
}

func TestDrain(t *testing.T) {
	workSize := 100
	q := New(func(*int) { time.Sleep(time.Nanosecond) })

	for i := 0; i < workSize; i++ {
		i := i
		q.EnQ(&i)
		q.Drain()
	}
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
	q := New(func(v *int) { fmt.Println(*v) })

	for i := 0; i < 100000; i++ {
		i := i
		q.EnQ(&i)
	}
}
