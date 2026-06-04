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

package commits

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

func TestQueryCommits_WithFirstCommitCid_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		// Result CIDs are hardcoded because template placeholders are not
		// resolved inside Request.Results.
		// See https://github.com/sourcenetwork/defradb/issues/4745.
		MultiplierExcludes: []string{multiplier.SignedDocs},
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			&action.Request{
				Request: `query {
						_commits(
							cid: "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu"
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithFirstCommitCidForFieldCommit_ShouldSucceed(t *testing.T) {
	// cid is for a field commit, see TestQueryCommitsWithDocIDAndFieldId
	test := testUtils.TestCase{
		// Result CIDs are hardcoded because template placeholders are not
		// resolved inside Request.Results.
		// See https://github.com/sourcenetwork/defradb/issues/4745.
		MultiplierExcludes: []string{multiplier.SignedDocs},
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"name": "Johnn"
				}`,
			},
			&action.Request{
				Request: `query {
						_commits(
							cid: "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm"
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithInvalidCid(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(cid: "fhbnjfahfhfhanfhga") {
							cid
							height
							delta
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
				ExpectedError: "invalid cid",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithInvalidShortCid(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(cid: "bafybeidfhbnjfahfhfhanfhga") {
							cid
							height
							delta
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
				ExpectedError: "invalid cid",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithUnknownCid(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(cid: "bafybeid57gpbwi4i6bg7g35hhhhhhhhhhhhhhhhhhhhhhhdoesnotexist") {
							cid
							height
							delta
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
				ExpectedError: "cid either does not exist or belong to document",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_MultipleCidsDifferentDocs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "Fred",
					"age":  21,
				},
			},
			&action.Request{
				Request: `query {
						_commits(
							cid: ["bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu", "bafyreibrfk6bijtty4vemejailnzvwpsclvyw47mossqgk547uvvawzkc4"]
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
						},
						{
							"cid": "bafyreibrfk6bijtty4vemejailnzvwpsclvyw47mossqgk547uvvawzkc4",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_MultipleCidsDifferentDocs_NoAccessToSecondCid(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy: `
description: a test policy which marks a collection in a database as a resource
name: test
resources:
- name: users
  permissions:
  - name: delete
  - expr: reader
    name: read
  - name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: reader
    types:
    - actor
`,
			},
			&action.AddCollection{
				SDL: `
				type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
					name: String
					age: Int
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "Fred",
					"age":  21,
				},
			},
			&action.Request{
				Request: `query {
						_commits(
							cid: ["bafyreib7pek5q7pweyljoylz7qusofwls3wnnzik6j5ibo3pe77fz4iw3e", "bafyreig2bqffvdrxc3bzot5m6bcm6qssybzgof6xdllxuy6cg7rtmpuwjy"]
						) {
							cid
						}
					}`,
				Results: map[string]any{
					// The second commit is not readable as it is not public and we are querying without
					// an identity
					"_commits": []map[string]any{
						{
							"cid": "bafyreib7pek5q7pweyljoylz7qusofwls3wnnzik6j5ibo3pe77fz4iw3e",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_MultipleCidsSameDoc(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			&action.Request{
				Request: `query {
						_commits(
							cid: ["bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu", "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm"]
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
						},
						{
							"cid": "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_ListOfOne(t *testing.T) {
	test := testUtils.TestCase{
		// Result CIDs are hardcoded because template placeholders are not
		// resolved inside Request.Results.
		// See https://github.com/sourcenetwork/defradb/issues/4745.
		MultiplierExcludes: []string{multiplier.SignedDocs},
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(
							cid: ["bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu"]
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
