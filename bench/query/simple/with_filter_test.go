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

	"github.com/sourcenetwork/defradb/bench/fixtures"
)

var (
	userSimpleWithFilterQuery1 = `
	query {
		User(filter: {Age: {_gt: 10}}) {
			_key
			Name
			Age
			Points
			Verified
		}
	}
	`

	userSimpleWithFilterQuery2 = `
	query {
		User(filter: {Age: {_gt: 50}}) {
			_key
			Name
			Age
			Points
			Verified
		}
	}
	`

	userSimpleWithFilterQuery3 = `
	query {
		User(filter: {Age: {_gt: 90}}) {
			_key
			Name
			Age
			Points
			Verified
		}
	}
	`

	userSimpleWithFilterQuery4 = `
	query {
		User(filter: {Points: {_gt: 10000}}) {
			_key
			Name
			Age
			Points
			Verified
		}
	}
	`

	userSimpleWithFilterQuery5 = `
	query {
		User(filter: {Verified: {_eq: true}}) {
			_key
			Name
			Age
			Points
			Verified
		}
	}
	`

	userSimpleWithFilterQuery6 = `
	query {
		User(filter: {Age: {_gt: 10}}) {
			_key
			Name
			Age
			Points
			Verified
		}
	}
	`

	userSimpleWithFilterQuery7 = `
	query {
		User(filter: {Age: {_gt: 50}}) {
			_key
			Age
		}
	}
	`

	userSimpleWithFilterQuery8 = `
	query {
		User(filter: {Age: {_gt: 90}}) {
			_key
			Age
		}
	}
	`

	userSimpleWithFilterQuery9 = `
	query {
		User(filter: {Points: {_gt: 10000}}) {
			_key
			Name
		}
	}
	`

	userSimpleWithFilterQuery10 = `
	query {
		User(filter: {Verified: {_eq: true}}) {
			_key
		}
	}
	`
)

func Benchmark_Query_UserSimple_Query1_WithFilter_Sync_1(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1, userSimpleWithFilterQuery1, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query1_WithFilter_Sync_10(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 10, userSimpleWithFilterQuery1, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query1_WithFilter_Sync_100(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 100, userSimpleWithFilterQuery1, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query1_WithFilter_Sync_1000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1000, userSimpleWithFilterQuery1, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query1_WithFilter_Sync_10000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 10000, userSimpleWithFilterQuery1, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query2_WithFilter_Sync_1000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1000, userSimpleWithFilterQuery2, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query2_WithFilter_Sync_10000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 10000, userSimpleWithFilterQuery2, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query3_WithFilter_Sync_1000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1000, userSimpleWithFilterQuery3, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query3_WithFilter_Sync_10000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 10000, userSimpleWithFilterQuery3, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query4_WithFilter_Sync_1000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1000, userSimpleWithFilterQuery4, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query4_WithFilter_Sync_10000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 10000, userSimpleWithFilterQuery4, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query5_WithFilter_Sync_1000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1000, userSimpleWithFilterQuery5, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query5_WithFilter_Sync_10000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 10000, userSimpleWithFilterQuery5, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query6_WithFilter_Sync_1000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1000, userSimpleWithFilterQuery6, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query6_WithFilter_Sync_10000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 10000, userSimpleWithFilterQuery6, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query7_WithFilter_Sync_1000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1000, userSimpleWithFilterQuery7, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query7_WithFilter_Sync_10000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 10000, userSimpleWithFilterQuery7, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query8_WithFilter_Sync_1000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1000, userSimpleWithFilterQuery8, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query8_WithFilter_Sync_10000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 10000, userSimpleWithFilterQuery8, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query9_WithFilter_Sync_1000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1000, userSimpleWithFilterQuery9, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query9_WithFilter_Sync_10000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 10000, userSimpleWithFilterQuery9, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query10_WithFilter_Sync_1000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 1000, userSimpleWithFilterQuery10, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Query_UserSimple_Query10_WithFilter_Sync_10000(b *testing.B) {
	ctx := context.Background()
	err := runQueryBenchGet(b, ctx, fixtures.ForSchema(ctx, "user_simple"), 10000, userSimpleWithFilterQuery10, false)
	if err != nil {
		b.Fatal(err)
	}
}
