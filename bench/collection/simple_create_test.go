package collection

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/bench/fixtures"
)

func Benchmark_Collection_UserSimple_Create_Sync_0_1(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchCreate(b, fixtures.WithSchema(ctx, "user_simple"), 0, 1, true)
	if err != nil {
		b.Fatal(err)
	}
}
