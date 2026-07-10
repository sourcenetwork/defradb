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

package sync

import (
	"testing"
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// A per-request block-sync timeout that is too small for any block fetch to complete aborts the
// DAG fetch, so the document does not materialise on the receiver. A generous per-request timeout
// on the same setup syncs it cleanly. Together these exercise the SyncDocuments BlockSyncTimeout
// option end-to-end: the option threads through to loadBlockLinks and actually bounds the
// per-block fetch (the generous case is the control that rules out an unrelated sync failure).
func TestDocSync_PerRequestBlockSyncTimeout_BoundsBlockFetch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			// A one-nanosecond per-block budget cannot be met, so the block fetch times out and the
			// document is not synced to node 1.
			testUtils.SyncDocs{
				NodeID:           1,
				CollectionID:     0,
				DocIDs:           []int{0},
				SourceNodes:      []int{0},
				BlockSyncTimeout: immutable.Some(time.Nanosecond),
			},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
			// A generous per-block budget on the same document syncs it through.
			testUtils.SyncDocs{
				NodeID:           1,
				CollectionID:     0,
				DocIDs:           []int{0},
				SourceNodes:      []int{0},
				BlockSyncTimeout: immutable.Some(30 * time.Second),
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"Name": "John"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
