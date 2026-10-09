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

package replicator

import (
	"fmt"
	"testing"
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// Backfilling more documents than the event buffer can hold must complete without
// blocking event delivery or exhausting libp2p stream resources.
func TestP2POneToOneReplicator_BackfillSyncsAllExistingDocs(t *testing.T) {
	actions := []any{
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
	}

	// Add many docs on the source node BEFORE configuring the
	// replicator. This is the case the test is about.
	const docCount = 200
	for i := range docCount {
		actions = append(actions, &action.AddDoc{
			NodeID: immutable.Some(0),
			Doc: fmt.Sprintf(`{
				"Name": "user-%d",
				"Age": %d
			}`, i, i),
		})
	}

	actions = append(actions,
		testUtils.AddReplicator{
			SourceNodeID: 0,
			TargetNodeID: 1,
			Timeout:      30 * time.Second,
		},
		testUtils.WaitForSync{},
		&action.Request{
			NodeID: immutable.Some(1),
			Request: `query {
				COUNT(Users: {})
			}`,
			Results: map[string]any{
				"COUNT": docCount,
			},
		},
	)

	testUtils.ExecuteTestCase(t, testUtils.TestCase{
		Actions: actions,
		// Releases before v1.2.0 lack the stream cleanup required on both peers
		// during backfill.
		OldFirstNodeSupportedFromVersion: "v1.2.0",
		OldLastNodeSupportedFromVersion:  "v1.2.0",
	})
}
