package bench

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/bench/fixtures"
	"github.com/sourcenetwork/defradb/db"
)

func Benchmark_Collection_UserSimpleOne_Read_1(b *testing.B) {
	ctx := context.Background()
	runCollectionBenchGet(b, fixtures.WithSchema(ctx, "user_simple"), 1)
}

func runCollectionBenchGet(b *testing.B, ctx fixtures.Context, count int) {
	_, err := db.NewDB(dbopts)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	// load fixtures

	// run benchmark
	for i := 0; i < b.N; i++ {
		for j := 0; j < count; j++ {

		}
	}
}
