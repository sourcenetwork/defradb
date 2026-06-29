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
			_docID
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
	err := runQueryParserBench(b, ctx, fixtures.ForCollection(ctx, "user_simple"), userSimpleQuery)
	if err != nil {
		b.Fatal(err)
	}
}
