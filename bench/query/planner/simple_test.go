package planner

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/bench/fixtures"
)

var (
	userSimpleQuery = `
	query {
		User {
			_key
			Name
			Age
			Points
			Verified
		}
	}
	`
)

func Benchmark_Planner_UserSimple_ParseQuery(b *testing.B) {
	ctx := context.Background()
	err := runQueryParserBench(b, ctx, fixtures.ForSchema(ctx, "user_simple"), userSimpleQuery)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Planner_UserSimple_MakePlan(b *testing.B) {
	ctx := context.Background()
	err := runMakePlanBench(b, ctx, fixtures.ForSchema(ctx, "user_simple"), userSimpleQuery)
	if err != nil {
		b.Fatal(err)
	}
}
