// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package collection

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/tests/bench/fixtures"
)

func Benchmark_Collection_UserSimple_Read_Sync_1_1(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 1, 1, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Sync_10_10(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 10, 10, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Sync_100_100(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 100, 100, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Sync_1000_1000(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 1000, 1000, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Sync_1000_1(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 1000, 1, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Sync_10000_1(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 10000, 1, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Sync_100000_1(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 100000, 1, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Sync_1000_10(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 1000, 10, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Sync_1000_100(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 1000, 100, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Async_1_1(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 1, 1, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Async_10_10(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 10, 10, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Async_100_100(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 100, 100, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Async_1000_1000(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 1000, 1000, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Async_1000_1(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 1000, 1, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Async_1000_10(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 1000, 10, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Async_1000_100(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 1000, 100, false)
	if err != nil {
		b.Fatal(err)
	}
}
