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

package branchable_collection

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestBranchableCollectionSync_OneNodeEmptyAnotherWithDocs_ShouldCopyAll(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User @branchable {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "John",
					"age":  30,
				},
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "Islam",
					"age":  25,
				},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.SyncBranchableCollection{
				NodeID: 1,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						name
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  30,
						},
						{
							"name": "Islam",
							"age":  25,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestBranchableCollectionSync_WithDifferentDocsOnBothNodes_ShouldSync(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User @branchable {
						name: String
					}
				`,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.AddDoc{
				NodeID: immutable.Some(1),
				DocMap: map[string]any{
					"name": "Islam",
				},
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "Andy",
				},
			},
			&action.AddDoc{
				NodeID: immutable.Some(1),
				DocMap: map[string]any{
					"name": "Fred",
				},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.SyncBranchableCollection{
				NodeID: 1,
			},
			&action.SyncBranchableCollection{
				NodeID: 0,
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
						{"name": "Andy"},
						{"name": "John"},
						{"name": "Islam"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestBranchableCollectionSync_ShouldNotSubscribe(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User @branchable {
						name: String
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.SyncBranchableCollection{
				NodeID: 1,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": gomega.HaveLen(1),
				},
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "Islam",
				},
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "Andy",
				},
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": gomega.HaveLen(1),
				},
			},
			&action.SyncBranchableCollection{
				NodeID: 1,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": gomega.HaveLen(3),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestBranchableCollectionSync_WithNonBranchableCollection_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.SyncBranchableCollection{
				NodeID:        0,
				ExpectedError: "collection is not branchable",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestBranchableCollectionSync_WithNonExistentCollection_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User @branchable {
						name: String
					}
				`,
			},
			&action.SyncBranchableCollection{
				NodeID:        0,
				CollectionID:  99, // Non-existent collection index
				ExpectedError: "index out of range",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
