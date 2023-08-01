// Copyright 2022 Democratized Data Foundation
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
)

var (
	// dockey will be replaced in the bench runner func
	userSimpleWithSingleLookupQuery = `
	query {
		User(dockey: "{{dockey}}") {
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
		fixtures.ForSchema(ctx, "user_simple"),
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
		fixtures.ForSchema(ctx, "user_simple"),
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
		fixtures.ForSchema(ctx, "user_simple"),
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
		fixtures.ForSchema(ctx, "user_simple"),
		1000,
		userSimpleWithSingleLookupQuery,
		false,
	)
	if err != nil {
		b.Fatal(err)
	}
}
