# `go test -bench` Reference

Writing and running benchmarks with the standard toolchain. Statistical comparison lives in [benchstat.md](./benchstat.md); profile reading lives in [pprof.md](./pprof.md).

## Commands

```bash
go test -run='^$' -bench=. -benchmem -count=10 ./... | tee bench.txt   # all benchmarks, skip unit tests, save for benchstat
go test -run='^$' -bench=BenchmarkEncode -benchmem -count=10 ./pkg/codec
go test -run='^$' -bench='BenchmarkEncode/size=4096' ./pkg/codec          # one sub-benchmark (slash-separated regexp per level)
go test -run='^$' -bench=. -benchtime=3s ./pkg/codec                       # longer runs per benchmark for slow operations
go test -run='^$' -bench=. -benchtime=100x ./pkg/codec                     # fixed iteration count
go test -run='^$' -bench=. -cpu=1,2,4,8 ./pkg/codec                        # sweep GOMAXPROCS (parallel benchmarks)
go test -run='^$' -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof ./pkg/codec
go test -c -o codec.test ./pkg/codec && ./codec.test -test.run='^$' -test.bench=. -test.count=10
```

| Flag | Purpose |
| --- | --- |
| `-bench=<regexp>` | Select benchmarks; `.` runs all |
| `-run='^$'` | Skip unit tests during a measurement session |
| `-benchmem` | Report `B/op` and `allocs/op` (same as `b.ReportAllocs()` in every benchmark) |
| `-count=N` | Repeat each benchmark N times — benchstat needs the samples |
| `-benchtime=3s` / `100x` | Minimum time per benchmark (default 1s) or fixed iteration count |
| `-cpu=1,2,4` | Run with each GOMAXPROCS value |
| `-cpuprofile` / `-memprofile` / `-trace` | Write profiles or an execution trace from the run |

## Reading a raw result line

```
BenchmarkEncode/size=64-8   5000000   230.5 ns/op   128 B/op   2 allocs/op
```

| Field                     | Meaning                                 |
| ------------------------- | --------------------------------------- |
| `BenchmarkEncode/size=64` | Benchmark name, then sub-benchmark name |
| `-8`                      | GOMAXPROCS during the run               |
| `5000000`                 | Iterations executed (`b.N`)             |
| `230.5 ns/op`             | Wall time per iteration                 |
| `128 B/op`                | Bytes heap-allocated per iteration      |
| `2 allocs/op`             | Heap allocations per iteration          |

## Sub-benchmarks by input size

```go
func BenchmarkEncode(b *testing.B) {
    for _, size := range []int{64, 256, 4096} {
        b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
            data := make([]byte, size) // setup per sub-benchmark, outside the timed loop
            b.ReportAllocs()
            for b.Loop() {
                Encode(data)
            }
        })
    }
}
```

Output rows are `BenchmarkEncode/size=64`, `BenchmarkEncode/size=256`, … — benchstat compares them by name, and `-col /size` turns the sizes into columns. A per-op cost that grows faster than the size points at a superlinear algorithm, which no single-size benchmark reveals.
