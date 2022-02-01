package query

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/bench/fixtures"
)

var (
	userSimpleWithLimitOffsetQuery = `
	query {
		User(limit: 10, offset: 5) {
			_key
			Name
			Age
			Points
			Verified
		}
	}
	`
)

func Benchmark_Query_UserSimple_Query_WithLimitOffset_Sync_1(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, fixtures.ForSchema(ctx, "user_simple"), 1, userSimpleWithLimitOffsetQuery, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query_WithLimitOffset_Sync_10(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, fixtures.ForSchema(ctx, "user_simple"), 10, userSimpleWithLimitOffsetQuery, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query_WithLimitOffset_Sync_100(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, fixtures.ForSchema(ctx, "user_simple"), 100, userSimpleWithLimitOffsetQuery, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query_WithLimitOffset_Sync_1000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, fixtures.ForSchema(ctx, "user_simple"), 1000, userSimpleWithLimitOffsetQuery, false)
	if err != nil {
		b.Fatal(err)
	}
}
