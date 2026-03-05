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

package collection

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/tests/bench/fixtures"
)

func Benchmark_Collection_UserSimple_AddMany_Sync_0_10(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchAddMany(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 0, 10, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimple_AddMany_Sync_0_100(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchAddMany(b, ctx, fixtures.ForCollection(ctx, "user_simple"), 0, 100, true)
	if err != nil {
		b.Fatal(err)
	}
}
