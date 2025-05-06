// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package collection

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/tests/bench/fixtures"
)

func Benchmark_Collection_UserSimple_Read_Sync_1_1(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1, 1, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Sync_10_10(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 10, 10, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Sync_100_100(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 100, 100, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Sync_1000_1000(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1000, 1000, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Sync_1000_1(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1000, 1, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Sync_10000_1(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 10000, 1, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Sync_100000_1(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 100000, 1, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Sync_1000_10(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1000, 10, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Sync_1000_100(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1000, 100, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Async_1_1(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1, 1, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Async_10_10(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 10, 10, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Async_100_100(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 100, 100, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Async_1000_1000(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1000, 1000, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Async_1000_1(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1000, 1, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Async_1000_10(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1000, 10, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Read_Async_1000_100(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1000, 100, false)
	if err != nil {
		b.Fatal(err)
	}
}
