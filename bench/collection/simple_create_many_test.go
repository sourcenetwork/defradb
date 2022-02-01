package collection

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/bench/fixtures"
)

func Benchmark_Collection_UserSimple_CreateMany_Sync_0_10(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchCreateMany(b, fixtures.ForSchema(ctx, "user_simple"), 0, 10, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_CreateMany_Sync_0_100(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchCreateMany(b, fixtures.ForSchema(ctx, "user_simple"), 0, 100, true)
	if err != nil {
		b.Fatal(err)
	}
}
