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

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

// Hardcoded collection-level commit CIDs; no template placeholder yet.
// See https://github.com/sourcenetwork/defradb/issues/4744.
var branchableCollectionCidExcludes = []string{multiplier.EncryptedDocs, multiplier.SignedDocs}

func TestQuerySimpleWithCidOfBranchableCollection_FirstCid(t *testing.T) {
	test := testUtils.TestCase{
		MultiplierExcludes: branchableCollectionCidExcludes,
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users @branchable {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"name": "Freddddd"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			&action.Request{
				Request: `query {
					Users (
							cid: "{{.CollectionCID0_0}}"
						) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithCidOfBranchableCollection_MiddleCid(t *testing.T) {
	test := testUtils.TestCase{
		MultiplierExcludes: branchableCollectionCidExcludes,
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users @branchable {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"name": "Freddddd"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			&action.Request{
				Request: `query {
					Users (
							cid: "{{.CollectionCID0_1}}"
						) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Freddddd",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithCidOfBranchableCollection_LastCid(t *testing.T) {
	test := testUtils.TestCase{
		MultiplierExcludes: branchableCollectionCidExcludes,
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users @branchable {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"name": "Freddddd"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			&action.Request{
				Request: `query {
					Users (
							cid: "{{.CollectionCID0}}"
						) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
						},
						{
							"name": "Freddddd",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
