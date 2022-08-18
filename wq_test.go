package wq_test

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/Karitham/wq"
)

func TestQueue(t *testing.T) {
	workerCount := runtime.NumCPU()
	workSize := 10000000

	q := wq.New[int](workerCount)
	// simulate some work so scheduler doesn't get stuck at all (not needed in an actual scenario)
	q.Run(func(*int) { time.Sleep(time.Nanosecond) })

	for i := 0; i < workSize; i++ {
		q.EnQ(&i)
	}
	q.Close()

	_, ok := q.DeQ(0)
	if ok {
		t.Error("queue not closed")
	}
}

func TestDrain(t *testing.T) {
	workerCount := runtime.NumCPU()
	workSize := 100
	q := wq.New[int](workerCount)
	// simulate some work so scheduler doesn't get stuck at all (not needed in an actual scenario)
	q.Run(func(*int) { time.Sleep(time.Nanosecond) })

	for i := 0; i < workSize; i++ {
		q.EnQ(&i)
		q.Drain()
	}
	q.Close()
}

func BenchmarkQueue(b *testing.B) {
	workerCount := runtime.NumCPU()
	q := wq.New[int](workerCount)

	// simulate some work so scheduler doesn't get stuck at all (not needed in an actual scenario)
	q.Run(func(*int) { time.Sleep(time.Nanosecond) })

	for i := 0; i < b.N; i++ {
		q.EnQ(&i)
	}
}

func BenchmarkChans(b *testing.B) {
	workerCount := runtime.NumCPU()
	c := make(chan int, workerCount)
	w := func() {
		for _, ok := <-c; ok; _, ok = <-c {
			// simulate some work so scheduler doesn't get stuck at all
			time.Sleep(time.Nanosecond)
		}
	}

	for i := 0; i < workerCount; i++ {
		go w()
	}

	for i := 0; i < b.N; i++ {
		c <- i
	}
}

func TestV(t *testing.T) {
	workerCount := runtime.NumCPU()
	q := wq.New[int](workerCount)

	q.Run(func(v *int) { fmt.Println(*v) })

	for i := 0; i < 100000; i++ {
		q.EnQ(&i)
	}
}
