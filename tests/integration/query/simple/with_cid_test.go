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

func TestQuerySimpleWithInvalidCid(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					Users (cid: "any non-nil string value - this will be ignored") {
						Name
					}
				}`,
				ExpectedError: "invalid cid: selected encoding not supported",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithCid(t *testing.T) {
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
				Doc: `{
					"name": "John"
				}`,
			},
			&action.Request{
				Request: `query {
					Users (
							cid: "bafyreifldhofx6cwi6ashk24rcefsuiqje5a2rziwcyte54z27wmgv4pey"
						) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_UnknownCid(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.Request{
				Request: `query {
					Users (
							cid: "bafyreifldhofx6cwi6ashk24rcefsuiqje5a2rziwcyte54z27wmgv4pey"
						) {
						name
					}
				}`,
				ExpectedError: "seek failed: (version fetcher) failed to get block in blockstore: ipld: could not find",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithCid_MultipleDocs(t *testing.T) {
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
				Doc: `{
					"name": "John"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			&action.Request{
				Request: `query {
					Users (
							cid: "bafyreifldhofx6cwi6ashk24rcefsuiqje5a2rziwcyte54z27wmgv4pey"
						) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithCIDAndCounterAfterUpdate_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		// Accumulated CRDT fields (pncounter/pcounter) cannot be indexed.
		// https://github.com/sourcenetwork/defradb/issues/4439
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						counter: Int @crdt(type: pcounter)
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"counter": int64(1),
				},
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"counter": 1}`,
			},
			&action.Request{
				Request: `query {
					User(cid: "{{.CID0_0_1}}") {
						counter
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"counter": int64(2),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithCidAfterDeleteOperation_ShouldReturnUser(t *testing.T) {
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
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.DeleteDoc{
				DocID: 0,
			},
			&action.Request{
				Request: `query {
					Users (
						cid: "bafyreic2vrbl344kkc7h5d7e2hpnwvffta4ck73bvjs5acgjtvqubvvioe"
						showDeleted: true
					){
						name
						_deleted
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":     "John",
							"_deleted": true,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_ListOfOneCID(t *testing.T) {
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
				Doc: `{
					"name": "John"
				}`,
			},
			&action.Request{
				Request: `query {
					Users (
							cid: ["bafyreifldhofx6cwi6ashk24rcefsuiqje5a2rziwcyte54z27wmgv4pey"]
						) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_MultipleCIDs(t *testing.T) {
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
				Doc: `{
					"name": "John"
				}`,
			},
			&action.Request{
				Request: `query {
					Users (
							cid: ["bafyreifldhofx6cwi6ashk24rcefsuiqje5a2rziwcyte54z27wmgv4pey", "bafyreic2vrbl344kkc7h5d7e2hpnwvffta4ck73bvjs5acgjtvqubvvioe"]
						) {
						name
					}
				}`,
				ExpectedError: "querying by multiple cids is not yet supported",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
