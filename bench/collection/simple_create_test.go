package collection

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/bench/fixtures"
)

func Benchmark_Collection_UserSimple_Create_Sync_0_1(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchCreate(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 0, 1, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Create_Sync_0_10(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchCreate(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 0, 10, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Create_Sync_0_100(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchCreate(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 0, 100, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Create_Sync_0_1000(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchCreate(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 0, 1000, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Create_Async_0_1(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchCreate(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 0, 1, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Create_Async_0_100(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchCreate(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 0, 100, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Create_Async_0_1000(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchCreate(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 0, 1000, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_Create_Async_0_100000(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchCreate(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 0, 100000, false)
	if err != nil {
		b.Fatal(err)
	}
}
