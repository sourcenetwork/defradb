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
)

var (
	// The `docID` will be replaced in the bench runner func
	userSimpleWithSingleLookupQuery = `
	query {
		User(docID: "{{docID}}") {
			_docID
			Name
			Age
			Points
			Verified
		}
	}
	`
)

func Benchmark_Query_UserSimple_Query_WithSingleLookup_Sync_1(b *testing.B) {
	ctx := context.Background()
	err := RunQueryBenchGet(
		b,
		ctx,
		fixtures.ForCollection(ctx, "user_simple"),
		1,
		userSimpleWithSingleLookupQuery,
		false,
	)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query_WithSingleLookup_Sync_10(b *testing.B) {
	ctx := context.Background()
	err := RunQueryBenchGet(
		b,
		ctx,
		fixtures.ForCollection(ctx, "user_simple"),
		10,
		userSimpleWithSingleLookupQuery,
		false,
	)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query_WithSingleLookup_Sync_100(b *testing.B) {
	ctx := context.Background()
	err := RunQueryBenchGet(
		b,
		ctx,
		fixtures.ForCollection(ctx, "user_simple"),
		100,
		userSimpleWithSingleLookupQuery,
		false,
	)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query_WithSingleLookup_Sync_1000(b *testing.B) {
	ctx := context.Background()
	err := RunQueryBenchGet(
		b,
		ctx,
		fixtures.ForCollection(ctx, "user_simple"),
		1000,
		userSimpleWithSingleLookupQuery,
		false,
	)
	if err != nil {
		b.Fatal(err)
	}
}
