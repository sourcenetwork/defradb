// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

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
			_docID
			Name
			Age
			Points
			Verified
		}
	}
	`
)

func makeUserAgeIndexOption() fixtures.Option {
	return fixtures.OptionFieldDirective("User", "Age", "@index")
}

func Benchmark_Index_UserSimple_QueryWithFilterOnIndex_Sync_1(b *testing.B) {
	ctx := context.Background()
	err := query.RunQueryBenchGet(
		b,
		ctx,
		fixtures.ForCollection(ctx, "user_simple", makeUserAgeIndexOption()),
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
		fixtures.ForCollection(ctx, "user_simple", makeUserAgeIndexOption()),
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
		fixtures.ForCollection(ctx, "user_simple", makeUserAgeIndexOption()),
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
		fixtures.ForCollection(ctx, "user_simple", makeUserAgeIndexOption()),
		10000,
		userSimpleWithFilterQuery,
		false,
	)
	if err != nil {
		b.Fatal(err)
	}
}
