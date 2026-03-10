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

package field

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersionUpdatesAddFieldWithAddWithUpdateAfterCollectionUpdateAndVersionJoin(t *testing.T) {
	initialCollectionVersionID := "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu"
	updatedCollectionVersionID := "bafyreigvzkfdc4y2ppvvpmmdw3t7kv4nd5dgfh5jfytef3kbzem6po55zu"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John"
				}`,
			},
			// We want to make sure that this works across database versions, so we tell
			// the change detector to split here.
			&action.Request{
				Request: `query {
					Users {
						name
						_version {
							collectionVersionId
							fieldName
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_version": []map[string]any{
								{
									"collectionVersionId": initialCollectionVersionID,
									"fieldName":           "_C",
								},
							},
						},
					},
				},
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"email": "ih8oraclelicensing@netscape.net"
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						email
						_version {
							collectionVersionId
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":  "John",
							"email": "ih8oraclelicensing@netscape.net",
							"_version": []map[string]any{
								{
									// Update commit
									"collectionVersionId": updatedCollectionVersionID,
								},
								{
									// Create commit
									"collectionVersionId": initialCollectionVersionID,
								},
							},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddFieldWithAddWithUpdateAfterCollectionUpdateAndCommitQuery(t *testing.T) {
	initialCollectionVersionID := "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu"
	updatedCollectionVersionID := "bafyreigvzkfdc4y2ppvvpmmdw3t7kv4nd5dgfh5jfytef3kbzem6po55zu"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John"
				}`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"email": "ih8oraclelicensing@netscape.net"
				}`,
			},
			&action.Request{
				Request: `query {
					_commits (filter: {fieldName: {_eq: "_C"}}) {
						collectionVersionId
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							// Update commit
							"collectionVersionId": updatedCollectionVersionID,
						},
						{
							// Create commit
							"collectionVersionId": initialCollectionVersionID,
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
