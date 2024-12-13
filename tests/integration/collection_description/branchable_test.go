// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package collection_description

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColDescr_Branchable(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users @branchable {}
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionDescription{
					{
						ID:             1,
						Name:           immutable.Some("Users"),
						IsMaterialized: true,
						IsBranchable:   true,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColDescr_BranchableIfTrue(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users @branchable(if: true) {}
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionDescription{
					{
						ID:             1,
						Name:           immutable.Some("Users"),
						IsMaterialized: true,
						IsBranchable:   true,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColDescr_BranchableIfFalse(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users @branchable(if: false) {}
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionDescription{
					{
						ID:             1,
						Name:           immutable.Some("Users"),
						IsMaterialized: true,
						IsBranchable:   false,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
