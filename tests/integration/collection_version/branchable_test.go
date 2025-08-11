// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package collection_version

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersion_Branchable(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users @branchable {}
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsBranchable:   true,
						IsActive:       true,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersion_BranchableIfTrue(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users @branchable(if: true) {}
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsBranchable:   true,
						IsActive:       true,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersion_BranchableIfFalse(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users @branchable(if: false) {}
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsBranchable:   false,
						IsActive:       true,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
