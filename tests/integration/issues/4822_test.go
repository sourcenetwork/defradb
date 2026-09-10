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

package issues

import (
	"fmt"
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// Reproduces a P2P replicator backfill drop under load.
//
// Setup: source node has many pre-existing documents in a single
// collection. A replicator from source -> target is then configured. The
// expectation (and the comment in `TestP2POneToOneReplicatorDoesNotSyncExisting`)
// is that `AddReplicator` backfills every existing document. In practice
// the backfill path in `internal/db/p2p/replicator.go::pushHeadsForAllDocs`
// drops documents partway through under load.
//
// Likely culprit: `pushHeadsForDoc` opens a fresh libp2p
// request stream per doc with a 10s `networkRequestTimeout`. Under
// back-to-back load some streams fail (or hang past 10s), and the
// failure mode also wedges subsequent collections in the iteration.
func TestP2POneToOneReplicator_BackfillDropsExistingDocsUnderLoad(t *testing.T) {
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
		},
		testUtils.WaitForSync{},
		&action.Request{
			NodeID: immutable.Some(1),
			Request: `query {
				_count(Users: {})
			}`,
			Results: map[string]any{
				"_count": docCount,
			},
		},
	)

	testUtils.ExecuteTestCase(t, testUtils.TestCase{Actions: actions})
}
