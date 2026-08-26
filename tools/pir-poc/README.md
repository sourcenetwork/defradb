# Dense XOR PIR cross-language benchmark

This directory is an isolated Go port of the Dense XOR kernel used by the
DefraDB Rust PIR POC. It does not change DefraDB storage, networking, queries,
or production APIs.

The executable and the Rust `cross-language-dense` example use the same JSON
schema, deterministic table bytes, target sequence, workload sizes, warmup,
sample counts, and correctness checks. Each Dense XOR query uses fresh
cryptographic randomness. The reported server time is **total work**: the sum
of sequential evaluation time across all replicas, not parallel wall time.
Client phases are repeated for at least 10 ms and server paths for at least 50
ms, then divided by the repetition count. This avoids zero-duration micro
results and reports warmed steady-state server work. The benchmark cleans the
Go heap before server timing because client and server are separate processes
in deployment.

The four measured paths are:

- direct retrieval, as a public lower bound;
- 100 visible candidates, the decoy baseline;
- Dense XOR with two non-colluding replicas;
- Dense XOR with three replicas, private if any one replica does not collude.

The public and decoy paths send 8-byte ordinals. They intentionally isolate
retrieval-kernel cost; tag hashing, directory construction and database lookup
are outside this language benchmark.

Run the small validation profile:

```console
go run ./tools/pir-poc -profile quick
```

Build first, then run the comparison profile so compilation is excluded:

```console
go build -o pir-poc-bench.exe ./tools/pir-poc
./pir-poc-bench.exe -profile full
```

The benchmark is deliberately a protocol-kernel comparison, not a claim that
Go and Rust services have identical HTTP, database, scheduler, or deployment
overhead.

## Measured comparison

Three alternating full run pairs on a Ryzen 7 3700X with Go 1.25.9 and Rust
1.94.0 found the following aggregate Dense XOR server work:

| Workload | Rust, 2 replicas | Go, 2 replicas | Rust, 3 replicas | Go, 3 replicas |
|---|---:|---:|---:|---:|
| 1,048,576 x 96 B, batch 1 | 15.04 ms | 19.49 ms (+29.6%) | 22.21 ms | 29.50 ms (+32.9%) |
| 65,536 x 2,008 B, batch 1 | 12.55 ms | 15.90 ms (+26.7%) | 18.40 ms | 23.38 ms (+27.1%) |
| 262,144 x 96 B, batch 16 | 58.74 ms | 78.60 ms (+33.8%) | 89.29 ms | 117.38 ms (+31.5%) |

Every client phase was below 0.51 ms. The Go implementation is viable, but the
current Rust kernel uses 21-25% less server time. Because aggregate server work
is the project objective, production should retain the Rust sidecar boundary
or first add an inlined SIMD/assembly XOR path and shared-row batch traversal
to Go. Wire sizes and privacy properties are identical across languages.
