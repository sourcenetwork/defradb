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

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// TestIndexBackfill_AfterSuccessfulBuild_ListsNoAction verifies that a successful index build
// completes (and so deletes) its action record, leaving nothing in ListActions.
func TestIndexBackfill_AfterSuccessfulBuild_ListsNoAction(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			// A synchronous build registers and completes its action before returning.
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "name",
			},
			&action.ListActions{
				ExpectedInfo: []client.ActionExecution{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
