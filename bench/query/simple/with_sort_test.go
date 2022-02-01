package query

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/bench/fixtures"
)

var (
	userSimpleWithSortQuery = `
	query {
		User(order: {Age: ASC}) {
			_key
			Name
			Age
			Points
			Verified
		}
	}
	`
)

func Benchmark_Query_UserSimple_Query_WithSort_Sync_1(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1, userSimpleWithSortQuery, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query_WithSort_Sync_10(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 10, userSimpleWithSortQuery, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query_WithSort_Sync_100(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 100, userSimpleWithSortQuery, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query_WithSort_Sync_1000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1000, userSimpleWithSortQuery, false)
	if err != nil {
		b.Fatal(err)
	}
}
