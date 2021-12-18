package read

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/bench/fixtures"
)

func Benchmark_Collection_UserSimpleOne_Read_Sync_1_1(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1, 1, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Sync_10_10(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 10, 10, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Sync_100_100(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 100, 100, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Sync_1000_1000(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1000, 1000, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Sync_1000_1(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1000, 1, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Sync_1000_10(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1000, 10, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Sync_1000_100(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1000, 100, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Async_1_1(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1, 1, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Async_10_10(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 10, 10, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Async_100_100(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 100, 100, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Async_1000_1000(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1000, 1000, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_1000_1_Async(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1000, 1, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_1000_10_Async(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1000, 10, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_1000_100_Async(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1000, 100, false)
	if err != nil {
		b.Fatal(err)
	}
}
