# Benchmark ACP

## DefraDB Commit
- This report was done on commit: `#10e0f38f`


## Goals
In the report we will benchmark the following DefraDB ACP Operations:
- Registering Document(s) With Local ACP
- Checking Document Access With Local ACP
- Checking If Document Is Registered With Local ACP

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
- For the initial setup we have a helper function below that will get a acp instance started.

```go
// newLocalACPSetup will setup localACP instance in memory if inMem is true and a persistent store otherwise.
// Additionally it will also start the acp instance.
// The caller is responsible to call `Close()` on the returned [acp.ACP] instance.
func newLocalACPSetup(b *testing.B, inMem bool) acp.ACP {
	ctx := context.Background()
	localACP := acp.NewLocalACP()

	if inMem {
		localACP.Init(ctx, "")
	} else {
		acpPath := b.TempDir()
		localACP.Init(ctx, acpPath)
	}

	err := localACP.Start(ctx)
	require.Nil(b, err)

	return localACP
}
```

- Then we have a helper to reset the acp state and load our desired policy.
```go
// resetLocalACPKeepPolicy resets the local acp instance then adds our desired policy back.
func resetLocalACPKeepPolicy(b *testing.B, ctx context.Context, localACP acp.ACP) {
	resetErr := localACP.ResetState(ctx)
	require.Nil(b, resetErr)

	policyID, errAddPolicy := localACP.AddPolicy(
		ctx,
		identity1,
		validPolicy,
	)
	require.Nil(b, errAddPolicy)
	require.Equal(
		b,
		validPolicyID,
		policyID,
	)
}
```

- Then we also have a helper to run Register call X times that we will use in benchmarking below.
```go
func registerXDocObjects(b *testing.B, ctx context.Context, count int, localACP acp.ACP) {
	for index := 0; index < count; index++ {
		err := localACP.RegisterDocObject(
			ctx,
			identity1,
			validPolicyID,
			"users",
			strconv.Itoa(index),
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}
```


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

BenchmarkACPRegister/scale=256,inMem=true-22       14     81.2 ms/op     62.8 MB/op       1M allocs/op
BenchmarkACPRegister/scale=512,inMem=true-22        4      283 ms/op      229 MB/op       4M allocs/op
BenchmarkACPRegister/scale=1024,inMem=true-22       1      1.07 s/op      871 MB/op      15M allocs/op
BenchmarkACPRegister/scale=2048,inMem=true-22       1      4.00 s/op     3.41 GB/op      60M allocs/op
BenchmarkACPRegister/scale=4096,inMem=true-22       1      16.1 s/op     13.5 GB/op     238M allocs/op
BenchmarkACPRegister/scale=8192,inMem=true-22       1      68.9 s/op     53.9 GB/op     946M allocs/op

BenchmarkACPRegister/scale=256,inMem=false-22      12     87.6 ms/op     90.0 MB/op       1M allocs/op
BenchmarkACPRegister/scale=512,inMem=false-22       4      324 ms/op      334 MB/op       5M allocs/op
BenchmarkACPRegister/scale=1024,inMem=false-22      1      1.24 s/op     1.28 GB/op      18M allocs/op
BenchmarkACPRegister/scale=2048,inMem=false-22      1      4.90 s/op     5.04 GB/op      69M allocs/op
BenchmarkACPRegister/scale=4096,inMem=false-22      1      21.3 s/op     20.0 GB/op     272M allocs/op
BenchmarkACPRegister/scale=8192,inMem=false-22      1      78.9 s/op     80.5 GB/op       1G allocs/op

PASS
ok      github.com/sourcenetwork/defradb/tests/bench/acp        204.853s
```

Observation:
- When the scale (number of doc objects registered) doubles (2x), the operations get 4x times slower.
- When the scale (number of doc objects registered) doubles (2x), the memory per operation increases by 4x.
- The scaling is similar for both In-memory and persistent acp.


## Benchmarking Local ACP Document Access Check

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

### Benchmarking with `[1, 2, 4, 8, 16, 32, 64, 128, 256]` registered document objects.

When we try running the above benchmarks with this scale of registered documents `[256, 512, 1024, 2048, 4096, 8192]`,
the benchmarks timeout because of the large `b.N` value due to setup taking much longer than actual bench operation.

So then minimized the scales to `[1, 2, 4, 8, 16, 32, 64, 128, 256, 512]` to see what we get, and only run In-memory
because even with this the tests took so long (look at how high the `b.N` values are). So long that had to Ctrl-C
right after `128` document benchmark result was printed.

Running Benchmark:
```sh
LOG_LEVEL=error humanbench go test -bench=. -benchmem -benchtime 1s
```

Results:
```sh
goos: linux
goarch: amd64
pkg: github.com/sourcenetwork/defradb/tests/bench/acp
cpu: Intel(R) Core(TM) Ultra 9 185H

BenchmarkACPCheckDocAccess/scale=1,inMem=true-22        7456      162 µs/op     71.8 kB/op      2k allocs/op
BenchmarkACPCheckDocAccess/scale=2,inMem=true-22       13945     82.6 µs/op     34.6 kB/op     768 allocs/op
BenchmarkACPCheckDocAccess/scale=4,inMem=true-22       14912     80.4 µs/op     34.6 kB/op     768 allocs/op
BenchmarkACPCheckDocAccess/scale=8,inMem=true-22       14467     83.2 µs/op     34.6 kB/op     768 allocs/op
BenchmarkACPCheckDocAccess/scale=16,inMem=true-22      13718     86.1 µs/op     34.6 kB/op     768 allocs/op
BenchmarkACPCheckDocAccess/scale=32,inMem=true-22      12457     97.2 µs/op     34.6 kB/op     768 allocs/op
BenchmarkACPCheckDocAccess/scale=64,inMem=true-22      10000      109 µs/op     34.7 kB/op     769 allocs/op
BenchmarkACPCheckDocAccess/scale=128,inMem=true-22      9360      120 µs/op     34.9 kB/op     771 allocs/op
```

Note: The `b.N` value reached almost 15K (14912 for scale 4) that means it is setting up acp state where
4 documents are setup 14912 times and then running the `CheckDocAccess` call and taking the average of that.

### Fixing the `CheckDocAccess` benchmarks by setting `b.N` to `10`:

Since the tests were taking too long we fixed the `b.N` to `10` and ran again:

```sh
LOG_LEVEL=error humanbench go test -bench=. -benchmem -benchtime=10x
```

Results:
```sh
goos: linux
goarch: amd64
pkg: github.com/sourcenetwork/defradb/tests/bench/acp
cpu: Intel(R) Core(TM) Ultra 9 185H

BenchmarkACPCheckDocAccess/scale=1,inMem=true-22        10      150 µs/op     71.8 kB/op      2k allocs/op
BenchmarkACPCheckDocAccess/scale=2,inMem=true-22        10     82.0 µs/op     34.6 kB/op     768 allocs/op
BenchmarkACPCheckDocAccess/scale=4,inMem=true-22        10     81.2 µs/op     34.6 kB/op     768 allocs/op
BenchmarkACPCheckDocAccess/scale=8,inMem=true-22        10     80.9 µs/op     34.6 kB/op     768 allocs/op
BenchmarkACPCheckDocAccess/scale=16,inMem=true-22       10     83.8 µs/op     34.6 kB/op     768 allocs/op
BenchmarkACPCheckDocAccess/scale=32,inMem=true-22       10      118 µs/op     34.7 kB/op     768 allocs/op
BenchmarkACPCheckDocAccess/scale=64,inMem=true-22       10     94.6 µs/op     34.7 kB/op     770 allocs/op
BenchmarkACPCheckDocAccess/scale=128,inMem=true-22      10      118 µs/op     34.8 kB/op     770 allocs/op
BenchmarkACPCheckDocAccess/scale=256,inMem=true-22      10      101 µs/op     35.0 kB/op     772 allocs/op
BenchmarkACPCheckDocAccess/scale=512,inMem=true-22      10      115 µs/op     34.9 kB/op     772 allocs/op

BenchmarkACPCheckDocAccess/scale=1,inMem=false-22       10      146 µs/op     79.0 kB/op      2k allocs/op
BenchmarkACPCheckDocAccess/scale=2,inMem=false-22       10     83.5 µs/op     38.4 kB/op     782 allocs/op
BenchmarkACPCheckDocAccess/scale=4,inMem=false-22       10     85.2 µs/op     38.4 kB/op     782 allocs/op
BenchmarkACPCheckDocAccess/scale=8,inMem=false-22       10     96.5 µs/op     38.4 kB/op     782 allocs/op
BenchmarkACPCheckDocAccess/scale=16,inMem=false-22      10     85.5 µs/op     38.4 kB/op     782 allocs/op
BenchmarkACPCheckDocAccess/scale=32,inMem=false-22      10     86.5 µs/op     38.4 kB/op     782 allocs/op
BenchmarkACPCheckDocAccess/scale=64,inMem=false-22      10      120 µs/op     38.4 kB/op     782 allocs/op
BenchmarkACPCheckDocAccess/scale=128,inMem=false-22     10     99.1 µs/op     38.6 kB/op     784 allocs/op
BenchmarkACPCheckDocAccess/scale=256,inMem=false-22     10      164 µs/op     39.0 kB/op     786 allocs/op
BenchmarkACPCheckDocAccess/scale=512,inMem=false-22     10      114 µs/op     38.7 kB/op     786 allocs/op
PASS
ok      github.com/sourcenetwork/defradb/tests/bench/acp        10.197s

```

Fixing `b.N` to `10` made the entire benchmark take only around 10 seconds, and we can see the pattern already.

So just for completeness we then adjust back to our original scale `[256, 512, 1024, 2048, 4096, 8192]`.

Keeping the rest of the benchmark code the same we run the benchmark again:

```sh
LOG_LEVEL=error humanbench go test -bench=. -benchmem -benchtime=10x
```

Results:
```sh
goos: linux
goarch: amd64
pkg: github.com/sourcenetwork/defradb/tests/bench/acp
cpu: Intel(R) Core(TM) Ultra 9 185H

BenchmarkACPCheckDocAccess/scale=256,inMem=true-22       10     126 µs/op     34.9 kB/op     772 allocs/op
BenchmarkACPCheckDocAccess/scale=512,inMem=true-22       10     160 µs/op     34.9 kB/op     772 allocs/op
BenchmarkACPCheckDocAccess/scale=1024,inMem=true-22      10     138 µs/op     34.9 kB/op     772 allocs/op
BenchmarkACPCheckDocAccess/scale=2048,inMem=true-22      10     177 µs/op     34.9 kB/op     772 allocs/op
BenchmarkACPCheckDocAccess/scale=4096,inMem=true-22      10     171 µs/op     35.0 kB/op     773 allocs/op
BenchmarkACPCheckDocAccess/scale=8192,inMem=true-22      10     216 µs/op     36.2 kB/op     774 allocs/op

BenchmarkACPCheckDocAccess/scale=256,inMem=false-22      10     116 µs/op     38.7 kB/op     786 allocs/op
BenchmarkACPCheckDocAccess/scale=512,inMem=false-22      10     111 µs/op     38.7 kB/op     786 allocs/op
BenchmarkACPCheckDocAccess/scale=1024,inMem=false-22     10     121 µs/op     38.8 kB/op     786 allocs/op
BenchmarkACPCheckDocAccess/scale=2048,inMem=false-22     10     149 µs/op     39.0 kB/op     786 allocs/op
BenchmarkACPCheckDocAccess/scale=4096,inMem=false-22     10     209 µs/op     61.8 kB/op     849 allocs/op
BenchmarkACPCheckDocAccess/scale=8192,inMem=false-22     10     250 µs/op     66.6 kB/op     910 allocs/op
```

Observation:
- When the scale (number of documents registered) doubles (2x), the operation speed does not slow down too much,
    the difference is very negligible (performance can improve, but the scaling is under control as we expect).
- When the scale (number of documents registered) doubles (2x), the memory remains similar per operation as we expect.
- The scaling is similar for both In-memory and persistent acp.


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

BenchmarkACPIsDocRegistered/scale=256,inMem=true-22       10      249 µs/op      221 kB/op          4k allocs/op
BenchmarkACPIsDocRegistered/scale=512,inMem=true-22       10      484 µs/op      422 kB/op          8k allocs/op
BenchmarkACPIsDocRegistered/scale=1024,inMem=true-22      10      957 µs/op      836 kB/op         15k allocs/op
BenchmarkACPIsDocRegistered/scale=2048,inMem=true-22      10     2.42 ms/op     1.64 MB/op         29k allocs/op
BenchmarkACPIsDocRegistered/scale=4096,inMem=true-22      10     3.98 ms/op     3.32 MB/op         58k allocs/op
BenchmarkACPIsDocRegistered/scale=8192,inMem=true-22      10     9.00 ms/op     6.53 MB/op        115k allocs/op

BenchmarkACPIsDocRegistered/scale=256,inMem=false-22      10      303 µs/op      323 kB/op          4k allocs/op
BenchmarkACPIsDocRegistered/scale=512,inMem=false-22      10      524 µs/op      623 kB/op          9k allocs/op
BenchmarkACPIsDocRegistered/scale=1024,inMem=false-22     10     1.26 ms/op     1.23 MB/op         17k allocs/op
BenchmarkACPIsDocRegistered/scale=2048,inMem=false-22     10     2.27 ms/op     2.43 MB/op         33k allocs/op
BenchmarkACPIsDocRegistered/scale=4096,inMem=false-22     10     5.38 ms/op     4.99 MB/op         67k allocs/op
BenchmarkACPIsDocRegistered/scale=8192,inMem=false-22     10     10.8 ms/op     9.85 MB/op         133k allocs/op
```

Observation:
- When the scale (number of doc objects registered) doubles (2x), the operations get 2x slower roughly.
- When the scale (number of doc objects registered) doubles (2x), the memory per operation increases by 2x.
- The scaling is similar for both In-memory and persistent acp.
