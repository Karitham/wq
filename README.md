# wq

Stupid lock-free work queue.

This was mostly an experiment, and hasn't been throughly tested, but should be equivalent or faster than a channel based one.

```benchmark
>> go test -benchmem -run=^$ -bench=. -benchtime=5s .
goos: linux
goarch: amd64
pkg: github.com/Karitham/wq
cpu: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
BenchmarkQueue-8        52272778               120.6 ns/op             0 B/op          0 allocs/op
BenchmarkChans-8        18827577               322.6 ns/op             0 B/op          0 allocs/op
PASS
ok      github.com/Karitham/wq  13.076s
```

It has the downside of requiring you to know your queue size ahead of time, which dictates the worker count.

It is also a busy queue, which works best if your producer are faster than your consumers.

## Usage

```go
q := wq.New(func(v *int) { fmt.Println(*v) })

for i := 0; i < 1000; i++ {
    i := i // copy (else it would be a pointer to the same value)
    q.EnQ(&i) // enqueue i
}

q.Wait() // wait for all workers to be done
```
