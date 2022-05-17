// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package planner

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/tests/bench/fixtures"
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
