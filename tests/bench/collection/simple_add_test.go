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

func Benchmark_Collection_UserSimple_Add_Sync_0_1(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchAdd(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 0, 1, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Add_Sync_0_10(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchAdd(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 0, 10, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Add_Sync_0_100(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchAdd(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 0, 100, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Add_Sync_0_1000(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchAdd(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 0, 1000, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Add_Async_0_1(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchAdd(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 0, 1, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Add_Async_0_100(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchAdd(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 0, 100, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Add_Async_0_1000(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchAdd(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 0, 1000, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Add_Async_0_10000(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchAdd(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 0, 10000, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Add_Async_0_100000(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchAdd(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 0, 100000, false)
	if err != nil {
		b.Fatal(err)
	}
}
