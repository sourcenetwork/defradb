// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package collection

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/tests/bench/fixtures"
)

func Benchmark_Collection_UserSimple_CreateMany_Sync_0_10(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchCreateMany(b, ctx, fixtures.ForSchema(ctx, "user_simple", nil), 0, 10, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_CreateMany_Sync_0_100(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchCreateMany(b, ctx, fixtures.ForSchema(ctx, "user_simple", nil), 0, 100, true)
	if err != nil {
		b.Fatal(err)
	}
}
