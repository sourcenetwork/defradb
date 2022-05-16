# DefraDB Benchmark Suite
This folder contains the DefraDB Benchmark Suite, its related code, sub packages, utilities, and data generators.

The goal of this suite is to provide an insight to DefraDBs performance, and to provide a quantitative approach to performance analysis and comparison. As such, the benchmark results should be used soley as a relative basis, and not concrete absolute values.

> Database benchmarking is a notorious complex issue to provide fair evaluations, that are void of contrived examples aimed to put the database "best foot forward".

## Workflow
The main benchmark suite should ideally be run on every PR as a "check" and alerts/warnings should be posted with relative performance deviates too much from some given "baseline".

There may be an additional `Nightly Suite` of benchmarks which would be run every night at XX:00 UTC which runs a more exhaustive benchmark, along with any tracing/analysis that would otherwise take too long to run within a PR check.

## Planned Tests
Here's a breakdown of the various benchmarks to run. This suite needs to cover *some* unit level benchmarks (testing specific components in isolation), to allow more fine-grained insights into the performance analysis. However, the majority of the suite should cover full integration benchmarks, primarily through the Query interface.

Some unit level benchmarks should be "cacheable" to avoid unnecessary overhead, we will try to identify which benchmarks *should* be cacheable.

### Workloads

#### Unit Benchmark
 - Underlying KV Engine `cacheable`
    - Read
    - Write
    - Mixed

#### Integration Benchmark