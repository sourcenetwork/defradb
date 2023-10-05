// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package query

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/tests/bench/fixtures"
	query "github.com/sourcenetwork/defradb/tests/bench/query/simple"
)

var (
	userSimpleWithFilterQuery = `
	query {
		User(filter: { Age: { _eq: 30 } }) {
			_key
			Name
			Age
			Points
			Verified
		}
	}
	`
)

func getOptions() *fixtures.Options {
	return fixtures.NewOptions().WithFieldDirective("User", "Age", "@index")
}

func Benchmark_Index_UserSimple_QueryWithFilterOnIndex_Sync_1(b *testing.B) {
	ctx := context.Background()
	err := query.RunQueryBenchGet(
		b,
		ctx,
		fixtures.ForSchema(ctx, "user_simple", getOptions()),
		1,
		userSimpleWithFilterQuery,
		false,
	)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Index_UserSimple_QueryWithFilterOnIndex_Sync_10(b *testing.B) {
	ctx := context.Background()
	err := query.RunQueryBenchGet(
		b,
		ctx,
		fixtures.ForSchema(ctx, "user_simple", getOptions()),
		10,
		userSimpleWithFilterQuery,
		false,
	)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Index_UserSimple_QueryWithFilterOnIndex_Sync_1000(b *testing.B) {
	ctx := context.Background()
	err := query.RunQueryBenchGet(
		b,
		ctx,
		fixtures.ForSchema(ctx, "user_simple", getOptions()),
		1000,
		userSimpleWithFilterQuery,
		false,
	)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Index_UserSimple_QueryWithFilterOnIndex_Sync_10000(b *testing.B) {
	ctx := context.Background()
	err := query.RunQueryBenchGet(
		b,
		ctx,
		fixtures.ForSchema(ctx, "user_simple", getOptions()),
		10000,
		userSimpleWithFilterQuery,
		false,
	)
	if err != nil {
		b.Fatal(err)
	}
}
