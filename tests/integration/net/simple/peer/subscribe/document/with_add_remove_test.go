// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package subscribe_test

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestP2PDocumentAddAndRemoveSingle(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToDocument{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
			testUtils.UnsubscribeToDocument{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Fred"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				// Fred has not been synced, as it was removed from the subscription set
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
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

func TestP2PDocumentAddAndRemoveMultiple(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Andy"
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToDocument{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
					state.NewColDocIndex(0, 1),
				},
			},
			testUtils.UnsubscribeToDocument{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
			testUtils.UpdateDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"name": "Bob"
				}`,
			},
			testUtils.UpdateDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				DocID:        1,
				Doc: `{
					"name": "Gillian"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Bob",
						},
						{
							"name": "Gillian",
						},
					},
				},
			},
			testUtils.Request{
				// Andy the User has not been synced, as the docID was removed from the subscription set.
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Gillian",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PDocumentAddSingleAndRemoveErroneous(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToDocument{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
			testUtils.UnsubscribeToDocument{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, testUtils.NonExistentDocID),
				},
				ExpectedError: "malformed document ID, missing either version or cid",
			},
			testUtils.UpdateDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"name": "Andy"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				// Andy has been synced, as the unsubscribe errored and should not have affected
				// the subscription to doc 0.
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Andy",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PDocumentAddSingleAndRemoveNone(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToDocument{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
			testUtils.UnsubscribeToDocument{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{},
			},
			testUtils.UpdateDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"name": "Andy"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				// Andy has been synced, as nothing was removed from the subscription set
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Andy",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
