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

func Benchmark_Query_UserSimple_Query_Sync_1(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(
		b,
		ctx,
		fixtures.ForSchema(ctx, "user_simple", nil),
		1,
		userSimpleQuery,
		false,
	)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query_Sync_10(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(
		b,
		ctx,
		fixtures.ForSchema(ctx, "user_simple", nil),
		10,
		userSimpleQuery,
		false,
	)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query_Sync_100(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(
		b,
		ctx,
		fixtures.ForSchema(ctx, "user_simple", nil),
		100,
		userSimpleQuery,
		false,
	)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query_Sync_1000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(
		b,
		ctx,
		fixtures.ForSchema(ctx, "user_simple", nil),
		1000,
		userSimpleQuery,
		false,
	)
	if err != nil {
		b.Fatal(err)
	}
}
