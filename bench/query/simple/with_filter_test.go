package query

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/bench/fixtures"
)

var (
	userSimpleWithFilterQuery = `
	query {
		User(filter: {Age: {_gt: 10}}) {
			_key
			Name
			Age
			Points
			Verified
		}
	}
	`
)

func Benchmark_Query_UserSimple_Query_WithFilter_Sync_1(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, fixtures.WithSchema(ctx, "user_simple"), 1, userSimpleWithFilterQuery, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query_WithFilter_Sync_10(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, fixtures.WithSchema(ctx, "user_simple"), 10, userSimpleWithFilterQuery, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query_WithFilter_Sync_100(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, fixtures.WithSchema(ctx, "user_simple"), 100, userSimpleWithFilterQuery, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query_WithFilter_Sync_1000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, fixtures.WithSchema(ctx, "user_simple"), 1000, userSimpleWithFilterQuery, false)
	if err != nil {
		b.Fatal(err)
	}
}
