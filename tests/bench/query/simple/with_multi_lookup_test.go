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
	// 10x `docID`s will be replaced in the bench runner func
	userSimpleWithMultiLookupQuery = `
	query {
		User(docID: ["{{docID}}", "{{docID}}", "{{docID}}", "{{docID}}", "{{docID}}", "{{docID}}", "{{docID}}", "{{docID}}", "{{docID}}", "{{docID}}"]) {
			_docID
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
	err := RunQueryBenchGet(
		b,
		ctx,
		fixtures.ForCollection(ctx, "user_simple"),
		10,
		userSimpleWithMultiLookupQuery,
		false,
	)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query_WithMultiLookup_Sync_100(b *testing.B) {
	ctx := context.Background()
	err := RunQueryBenchGet(
		b,
		ctx,
		fixtures.ForCollection(ctx, "user_simple"),
		100,
		userSimpleWithMultiLookupQuery,
		false,
	)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query_WithMultiLookup_Sync_1000(b *testing.B) {
	ctx := context.Background()
	err := RunQueryBenchGet(
		b,
		ctx,
		fixtures.ForCollection(ctx, "user_simple"),
		1000,
		userSimpleWithMultiLookupQuery,
		false,
	)
	if err != nil {
		b.Fatal(err)
	}
}
