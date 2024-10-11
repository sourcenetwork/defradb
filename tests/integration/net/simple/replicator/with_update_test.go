// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package replicator

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestP2POneToOneReplicatorUpdatesDocCreatedBeforeReplicatorConfig(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				// This document is created in first node before the replicator is set up.
				// Updates should be synced across nodes.
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.UpdateDoc{
				// Update John's Age on the first node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Age": 60
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(60),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2POneToOneReplicatorUpdatesDocCreatedBeforeReplicatorConfigWithNodesInversed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				// This document is created in second node before the replicator is set up.
				// Updates should be synced across nodes.
				NodeID: immutable.Some(1),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.ConfigureReplicator{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.UpdateDoc{
				// Update John's Age on the second node only, and allow the value to sync
				NodeID: immutable.Some(1),
				Doc: `{
					"Age": 60
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(60),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2POneToOneReplicator_ManyDocsUpdateWithTargetNodeTemporarilyOffline_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		SupportedDatabaseTypes: immutable.Some(
			[]testUtils.DatabaseType{
				// This test only supports file type databases since it requires the ability to
				// stop and start a node without losing data.
				testUtils.BadgerFileType,
			},
		),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.Close{
				NodeID: immutable.Some(1),
			},
			testUtils.CreateDoc{
				// Create John on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				// Create Fred on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "Fred",
					"Age": 22
				}`,
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				DocID:  0,
				Doc:    `{"Age": 22}`,
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				DocID:  1,
				Doc:    `{"Age": 23}`,
			},
			testUtils.Start{
				NodeID: immutable.Some(1),
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(23),
						},
						{
							"Age": int64(22),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2POneToOneReplicator_ManyDocsUpdateWithTargetNodeTemporarilyOfflineAfterCreate_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		SupportedDatabaseTypes: immutable.Some(
			[]testUtils.DatabaseType{
				// This test only supports file type databases since it requires the ability to
				// stop and start a node without losing data.
				testUtils.BadgerFileType,
			},
		),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.CreateDoc{
				// Create John on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				// Create Fred on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "Fred",
					"Age": 22
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Close{
				NodeID: immutable.Some(1),
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				DocID:  0,
				Doc:    `{"Age": 22}`,
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				DocID:  1,
				Doc:    `{"Age": 23}`,
			},
			testUtils.Start{
				NodeID: immutable.Some(1),
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(23),
						},
						{
							"Age": int64(22),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
