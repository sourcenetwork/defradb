# Benchmark ACP

## DefraDB Commit
- This report was done when acp core was bumped to: `v0.0.0-20250119220233-a3579dc36a39`
- The bump was done in commit: `e756de13e8891f051a997081400758f8b0579c18` [here](https://github.com/sourcenetwork/defradb/commit/e756de13e8891f051a997081400758f8b0579c18)


## Goals
In the report we will benchmark the following DefraDB ACP Operations:
- Registering Document(s) With Local ACP
- Checking Document Access With Local ACP
- Checking If Document Is Registered With Local ACP
- Comparing the results with the previous benchmarks done in `acp_bench_report_10e0f38f.md`

## Pre-Requisites

### Human Readable Result

- We will use [humanbench](https://github.com/kevinburke/humanbench) with `go bench` to make the results more readable.
To install, simply just do:
```
go install github.com/kevinburke/humanbench@latest
```

To run the benchmark we will then do:
```
LOG_LEVEL=error humanbench go test -bench=. -benchmem -benchtime 1s
```

### Fix `b.N` to 10 in cases with long setup times, and short operation times.
For some benchmarks where the setup took most of the time the `b.N` would go over `10k`, and for benchmarks
that would need the setup to register 10K documents before benchmarking a single operation would cause the
benchmarks to wastefully add `10K` documents `10k` times, which is a waste.

So for those instances we fix the `b.N` to `10`, which is still pretty accurate for our cases. For those
benchmarks our command would look like:

```
LOG_LEVEL=error humanbench go test -bench=. -benchmem -benchtime=10x
```

### Local ACP Setup
- The setup remains the same as our previous report: `acp_bench_report_10e0f38f.md`


## Benchmarking Local ACP Register Call

### Benchmark registering `[256, 512, 1024, 2048, 4096, 8192]` document objects.

Benchmarking Code:
```go
func BenchmarkACPRegister(b *testing.B) {
	for _, inMemoryOrPersistent := range []bool{true, false} {
		for _, scaleBy := range []int{256, 512, 1024, 2048, 4096, 8192} {
			b.Run(
				fmt.Sprintf("scale=%d,inMem=%t", scaleBy, inMemoryOrPersistent),
				func(b *testing.B) {
					localACP := newLocalACPSetup(b, inMemoryOrPersistent)
					defer localACP.Close()

					b.ResetTimer()
					for bNIndex := 0; bNIndex < b.N; bNIndex++ {
						// Since we need to re-initialize for every run use stop-start.
						b.StopTimer()
						ctx := context.Background()
						resetLocalACPKeepPolicy(b, ctx, localACP)

						b.StartTimer()
						registerXDocObjects(b, ctx, scaleBy, localACP)
					}
				},
			)
		}
	}
}
```

Note: The term `scale` refers to number of documents that are registered.

Run Benchmark:
```sh
LOG_LEVEL=error humanbench go test -bench=. -benchmem -benchtime 1s
```

Results:
```sh
goos: linux
goarch: amd64
pkg: github.com/sourcenetwork/defradb/tests/bench/acp
cpu: Intel(R) Core(TM) Ultra 9 185H

BenchmarkACPRegister/scale=256,inMem=true-22       57     21.8 ms/op     11.6 MB/op     213k allocs/op
BenchmarkACPRegister/scale=512,inMem=true-22       25     44.5 ms/op     23.1 MB/op     426k allocs/op
BenchmarkACPRegister/scale=1024,inMem=true-22      13     88.7 ms/op     46.3 MB/op     852k allocs/op
BenchmarkACPRegister/scale=2048,inMem=true-22       6      178 ms/op     92.6 MB/op       2M allocs/op
BenchmarkACPRegister/scale=4096,inMem=true-22       3      364 ms/op      185 MB/op       3M allocs/op
BenchmarkACPRegister/scale=8192,inMem=true-22       2      744 ms/op      371 MB/op       7M allocs/op

BenchmarkACPRegister/scale=256,inMem=false-22      62     16.2 ms/op     13.8 MB/op     223k allocs/op
BenchmarkACPRegister/scale=512,inMem=false-22      33     32.8 ms/op     27.6 MB/op     445k allocs/op
BenchmarkACPRegister/scale=1024,inMem=false-22     18     67.2 ms/op     55.4 MB/op     891k allocs/op
BenchmarkACPRegister/scale=2048,inMem=false-22      8      135 ms/op      111 MB/op       2M allocs/op
BenchmarkACPRegister/scale=4096,inMem=false-22      4      276 ms/op      226 MB/op       4M allocs/op
BenchmarkACPRegister/scale=8192,inMem=false-22      2      601 ms/op      476 MB/op       7M allocs/op

PASS
ok      github.com/sourcenetwork/defradb/tests/bench/acp        21.075s
```

Observation:
- When the scale (number of doc objects registered) doubles (2x), the operations get 2x times slower.
- When the scale (number of doc objects registered) doubles (2x), the memory per operation increases by 2x.
- The scaling is similar for both In-memory and persistent acp.

#### Improvement over previous results:
- Previously the operation and memory would be x4 more everytime the documents doubled. now it is x2.

- However there is overall improve of around 100x if we don't only look at the 2x scale magnitude improvement.

For example:
```sh
Old:
BenchmarkACPRegister/scale=8192,inMem=true-22       1      68.9 s/op     53.9 GB/op     946M allocs/op

New:
BenchmarkACPRegister/scale=8192,inMem=true-22       2      744 ms/op      371 MB/op       7M allocs/op
```

## Benchmarking Local ACP Document Access Check

### Benchmarking with `[256, 512, 1024, 2048, 4096, 8192]` registered document objects.

Benchmarking Code:
```go
func BenchmarkACPCheckDocAccess(b *testing.B) {
	for _, inMemoryOrPersistent := range []bool{true, false} {
		for _, scaleBy := range []int{256, 512, 1024, 2048, 4096, 8192} {
			b.Run(
				fmt.Sprintf("scale=%d,inMem=%t", scaleBy, inMemoryOrPersistent),
				func(b *testing.B) {
					localACP := newLocalACPSetup(b, inMemoryOrPersistent)
					defer localACP.Close()

					b.ResetTimer()
					for bNIndex := 0; bNIndex < b.N; bNIndex++ {
						// Since we need to re-initialize for every run use stop-start.
						b.StopTimer()
						ctx := context.Background()
						resetLocalACPKeepPolicy(b, ctx, localACP)
						registerXDocObjects(b, ctx, scaleBy, localACP)

						b.StartTimer()
						_, err := localACP.CheckDocAccess(
							ctx,
							acp.ReadPermission,
							identity1.DID,
							validPolicyID,
							"users",
							"1",
						)
						if err != nil {
							b.Fatal(err)
						}
					}
				},
			)
		}
	}
}
```

Note: The term `scale` refers to number of documents that are registered, before the `CheckDocAccess` call.


```sh
LOG_LEVEL=error humanbench go test -bench=. -benchmem -benchtime=10x
```

Results:
```sh
goos: linux
goarch: amd64
pkg: github.com/sourcenetwork/defradb/tests/bench/acp
cpu: Intel(R) Core(TM) Ultra 9 185H

BenchmarkACPCheckDocAccess/scale=256,inMem=true-22       10      101 µs/op     34.7 kB/op     769 allocs/op
BenchmarkACPCheckDocAccess/scale=512,inMem=true-22       10      102 µs/op     34.8 kB/op     770 allocs/op
BenchmarkACPCheckDocAccess/scale=1024,inMem=true-22      10      101 µs/op     34.9 kB/op     771 allocs/op
BenchmarkACPCheckDocAccess/scale=2048,inMem=true-22      10      101 µs/op     34.9 kB/op     771 allocs/op
BenchmarkACPCheckDocAccess/scale=4096,inMem=true-22      10      195 µs/op     34.9 kB/op     772 allocs/op
BenchmarkACPCheckDocAccess/scale=8192,inMem=true-22      10      102 µs/op     34.9 kB/op     772 allocs/op

BenchmarkACPCheckDocAccess/scale=256,inMem=false-22      10     99.4 µs/op     38.5 kB/op     784 allocs/op
BenchmarkACPCheckDocAccess/scale=512,inMem=false-22      10      104 µs/op     38.6 kB/op     784 allocs/op
BenchmarkACPCheckDocAccess/scale=1024,inMem=false-22     10      108 µs/op     38.7 kB/op     786 allocs/op
BenchmarkACPCheckDocAccess/scale=2048,inMem=false-22     10      108 µs/op     38.7 kB/op     786 allocs/op
BenchmarkACPCheckDocAccess/scale=4096,inMem=false-22     10      310 µs/op     39.0 kB/op     792 allocs/op
BenchmarkACPCheckDocAccess/scale=8192,inMem=false-22     10      127 µs/op     51.4 kB/op     853 allocs/op
```

Observation:
- When the scale (number of documents registered) doubles (2x), the operation speed does not slow down too much.
- When the scale (number of documents registered) doubles (2x), the memory remains similar per operation as we expect.
- The scaling is similar for both In-memory and persistent acp.

#### Improvement over previous results:
- The new and previous implementations behave very similarly. There is no significant bump up.


## Benchmarking Local ACP Checking If Object/Document Is Registered

### Benchmark using `[256, 512, 1024, 2048, 4096, 8192]` registered document objects.

Benchmarking Code:
```go
func BenchmarkACPIsDocRegistered(b *testing.B) {
	for _, inMemoryOrPersistent := range []bool{true, false} {
		for _, scaleBy := range []int{256, 512, 1024, 2048, 4096, 8192} {
			b.Run(
				fmt.Sprintf("scale=%d,inMem=%t", scaleBy, inMemoryOrPersistent),
				func(b *testing.B) {
					localACP := newLocalACPSetup(b, inMemoryOrPersistent)
					defer localACP.Close()

					b.ResetTimer()
					for bNIndex := 0; bNIndex < b.N; bNIndex++ {
						// Since we need to re-initialize for every run use stop-start.
						b.StopTimer()
						ctx := context.Background()
						resetLocalACPKeepPolicy(b, ctx, localACP)
						registerXDocObjects(b, ctx, scaleBy, localACP)

						b.StartTimer()
						_, err := localACP.IsDocRegistered(ctx, validPolicyID, "users", "1")
						if err != nil {
							b.Fatal(err)
						}
					}
				},
			)
		}
	}
}
```

Note: The term `scale` refers to number of documents that are registered, before the `IsDocRegistered` call.

Run Benchmark:
```sh
LOG_LEVEL=error humanbench go test -bench=. -benchmem -benchtime=10x
```

Results:
```sh
goos: linux
goarch: amd64
pkg: github.com/sourcenetwork/defradb/tests/bench/acp
cpu: Intel(R) Core(TM) Ultra 9 185H

BenchmarkACPIsDocRegistered/scale=256,inMem=true-22       10     54.6 µs/op     20.9 kB/op     395 allocs/op
BenchmarkACPIsDocRegistered/scale=512,inMem=true-22       10     49.1 µs/op     20.9 kB/op     395 allocs/op
BenchmarkACPIsDocRegistered/scale=1024,inMem=true-22      10     44.3 µs/op     20.9 kB/op     395 allocs/op
BenchmarkACPIsDocRegistered/scale=2048,inMem=true-22      10     50.0 µs/op     20.9 kB/op     395 allocs/op
BenchmarkACPIsDocRegistered/scale=4096,inMem=true-22      10     47.0 µs/op     20.9 kB/op     395 allocs/op
BenchmarkACPIsDocRegistered/scale=8192,inMem=true-22      10     50.1 µs/op     20.9 kB/op     395 allocs/op

BenchmarkACPIsDocRegistered/scale=256,inMem=false-22      10     49.1 µs/op     25.3 kB/op     413 allocs/op
BenchmarkACPIsDocRegistered/scale=512,inMem=false-22      10     42.7 µs/op     25.2 kB/op     413 allocs/op
BenchmarkACPIsDocRegistered/scale=1024,inMem=false-22     10     39.9 µs/op     25.2 kB/op     413 allocs/op
BenchmarkACPIsDocRegistered/scale=2048,inMem=false-22     10     44.1 µs/op     25.2 kB/op     413 allocs/op
BenchmarkACPIsDocRegistered/scale=4096,inMem=false-22     10      343 µs/op     25.6 kB/op     421 allocs/op
BenchmarkACPIsDocRegistered/scale=8192,inMem=false-22     10     80.9 µs/op     50.7 kB/op     506 allocs/op
```

Observation:
- When the scale (number of doc objects registered) doubles (2x), the operations get 2x slower roughly.
- When the scale (number of doc objects registered) doubles (2x), the memory per operation increases by 2x.
- The scaling is similar for both In-memory and persistent acp.

Observation:
- When the scale (number of doc objects registered) doubles (2x), the operations stay consistent.
- When the scale (number of doc objects registered) doubles (2x), the memory per operation stays consistent.
- The scaling is similar for both In-memory and persistent acp.

#### Improvement over previous results:
- Previously the operation and memory would be x2 more every time the documents doubled. Now it is constant.

- There is a massive improvement (specially in the in-memory implementation).

For example, in this case there is a 180x sped up per operation and 312x less memory per operation:
```sh
Old:
BenchmarkACPIsDocRegistered/scale=8192,inMem=true-22      10     9.00 ms/op     6.53 MB/op        115k allocs/op

New:
BenchmarkACPIsDocRegistered/scale=8192,inMem=true-22      10     50.1 µs/op     20.9 kB/op     395 allocs/op
```

## TLDR Comparisons of Old vs New:

Register:
```go
BenchmarkACPRegisterOld/scale=256,inMem=true-22       14     81.2 ms/op     62.8 MB/op       1M allocs/op
BenchmarkACPRegisterOld/scale=512,inMem=true-22        4      283 ms/op      229 MB/op       4M allocs/op
BenchmarkACPRegisterOld/scale=1024,inMem=true-22       1      1.07 s/op      871 MB/op      15M allocs/op
BenchmarkACPRegisterOld/scale=2048,inMem=true-22       1      4.00 s/op     3.41 GB/op      60M allocs/op
BenchmarkACPRegisterOld/scale=4096,inMem=true-22       1      16.1 s/op     13.5 GB/op     238M allocs/op
BenchmarkACPRegisterOld/scale=8192,inMem=true-22       1      68.9 s/op     53.9 GB/op     946M allocs/op

BenchmarkACPRegisterNew/scale=256,inMem=true-22       57     21.8 ms/op     11.6 MB/op     213k allocs/op
BenchmarkACPRegisterNew/scale=512,inMem=true-22       25     44.5 ms/op     23.1 MB/op     426k allocs/op
BenchmarkACPRegisterNew/scale=1024,inMem=true-22      13     88.7 ms/op     46.3 MB/op     852k allocs/op
BenchmarkACPRegisterNew/scale=2048,inMem=true-22       6      178 ms/op     92.6 MB/op       2M allocs/op
BenchmarkACPRegisterNew/scale=4096,inMem=true-22       3      364 ms/op      185 MB/op       3M allocs/op
BenchmarkACPRegisterNew/scale=8192,inMem=true-22       2      744 ms/op      371 MB/op       7M allocs/op
```


CheckDocAccess:

```go
BenchmarkACPCheckDocAccessOld/scale=256,inMem=true-22       10     126 µs/op     34.9 kB/op     772 allocs/op
BenchmarkACPCheckDocAccessOld/scale=512,inMem=true-22       10     160 µs/op     34.9 kB/op     772 allocs/op
BenchmarkACPCheckDocAccessOld/scale=1024,inMem=true-22      10     138 µs/op     34.9 kB/op     772 allocs/op
BenchmarkACPCheckDocAccessOld/scale=2048,inMem=true-22      10     177 µs/op     34.9 kB/op     772 allocs/op
BenchmarkACPCheckDocAccessOld/scale=4096,inMem=true-22      10     171 µs/op     35.0 kB/op     773 allocs/op
BenchmarkACPCheckDocAccessOld/scale=8192,inMem=true-22      10     216 µs/op     36.2 kB/op     774 allocs/op

BenchmarkACPCheckDocAccessNew/scale=256,inMem=true-22       10     101 µs/op     34.7 kB/op     769 allocs/op
BenchmarkACPCheckDocAccessNew/scale=512,inMem=true-22       10     102 µs/op     34.8 kB/op     770 allocs/op
BenchmarkACPCheckDocAccessNew/scale=1024,inMem=true-22      10     101 µs/op     34.9 kB/op     771 allocs/op
BenchmarkACPCheckDocAccessNew/scale=2048,inMem=true-22      10     101 µs/op     34.9 kB/op     771 allocs/op
BenchmarkACPCheckDocAccessNew/scale=4096,inMem=true-22      10     195 µs/op     34.9 kB/op     772 allocs/op
BenchmarkACPCheckDocAccessNew/scale=8192,inMem=true-22      10     102 µs/op     34.9 kB/op     772 allocs/op
```


IsDocRegistered:

```go
BenchmarkACPIsDocRegisteredOld/scale=256,inMem=true-22       10      249 µs/op      221 kB/op      4k allocs/op
BenchmarkACPIsDocRegisteredOld/scale=512,inMem=true-22       10      484 µs/op      422 kB/op      8k allocs/op
BenchmarkACPIsDocRegisteredOld/scale=1024,inMem=true-22      10      957 µs/op      836 kB/op     15k allocs/op
BenchmarkACPIsDocRegisteredOld/scale=2048,inMem=true-22      10     2.42 ms/op     1.64 MB/op     29k allocs/op
BenchmarkACPIsDocRegisteredOld/scale=4096,inMem=true-22      10     3.98 ms/op     3.32 MB/op     58k allocs/op
BenchmarkACPIsDocRegisteredOld/scale=8192,inMem=true-22      10     9.00 ms/op     6.53 MB/op    115k allocs/op

BenchmarkACPIsDocRegisteredNew/scale=256,inMem=true-22       10     54.6 µs/op     20.9 kB/op     395 allocs/op
BenchmarkACPIsDocRegisteredNew/scale=512,inMem=true-22       10     49.1 µs/op     20.9 kB/op     395 allocs/op
BenchmarkACPIsDocRegisteredNew/scale=1024,inMem=true-22      10     44.3 µs/op     20.9 kB/op     395 allocs/op
BenchmarkACPIsDocRegisteredNew/scale=2048,inMem=true-22      10     50.0 µs/op     20.9 kB/op     395 allocs/op
BenchmarkACPIsDocRegisteredNew/scale=4096,inMem=true-22      10     47.0 µs/op     20.9 kB/op     395 allocs/op
BenchmarkACPIsDocRegisteredNew/scale=8192,inMem=true-22      10     50.1 µs/op     20.9 kB/op     395 allocs/op
```
