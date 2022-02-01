package query

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/bench/fixtures"
)

var (
	// 10x dockey will be replaced in the bench runner func
	userSimpleWithMultiLookupQuery = `
	query {
		User(dockeys: ["{{dockey}}", "{{dockey}}", "{{dockey}}", "{{dockey}}", "{{dockey}}", "{{dockey}}", "{{dockey}}", "{{dockey}}", "{{dockey}}", "{{dockey}}"]) {
			_key
			Name
			Age
			Points
			Verified
		}
	}
	`
)

func Benchmark_Query_UserSimple_Query_WithMultiLookup_Sync_10(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, fixtures.WithSchema(ctx, "user_simple"), 10, userSimpleWithMultiLookupQuery, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query_WithMultiLookup_Sync_100(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, fixtures.WithSchema(ctx, "user_simple"), 100, userSimpleWithMultiLookupQuery, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query_WithMultiLookup_Sync_1000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, fixtures.WithSchema(ctx, "user_simple"), 1000, userSimpleWithMultiLookupQuery, false)
	if err != nil {
		b.Fatal(err)
	}
}
