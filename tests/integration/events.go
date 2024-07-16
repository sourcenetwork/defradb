// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

import (
	"context"
	"time"

	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"
)

// waitForUpdateEvents waits for all selected nodes to publish an update event to the local event bus.
//
// Expected document heads will be updated for all nodes that should receive network merges.
func waitForUpdateEvents(
	s *state,
	nodeID immutable.Option[int],
	collectionID int,
) {
	ctx, cancel := context.WithTimeout(s.ctx, subscriptionTimeout*10)
	defer cancel()

	for i := 0; i < len(s.nodes); i++ {
		if nodeID.HasValue() && nodeID.Value() != i {
			continue // node is not selected
		}

		var evt event.Update
		select {
		case msg := <-s.nodeEvents[i].update.Message():
			evt = msg.Data.(event.Update)

		case <-ctx.Done():
			require.Fail(s.t, "timeout waiting for update event")
		}

		// update the actual document heads
		s.actualDocHeads[i][evt.DocID] = evt.Cid

		// update the expected document heads of connected nodes
		for id := range s.nodeConnections[i] {
			if _, ok := s.actualDocHeads[id][evt.DocID]; ok {
				s.expectedDocHeads[id][evt.DocID] = evt.Cid
			}
		}
		// update the expected document heads of replicator sources
		for id := range s.nodeReplicatorTargets[i] {
			if _, ok := s.actualDocHeads[id][evt.DocID]; ok {
				s.expectedDocHeads[id][evt.DocID] = evt.Cid
			}
		}
		// update the expected document heads of replicator targets
		for id := range s.nodeReplicatorSources[i] {
			s.expectedDocHeads[id][evt.DocID] = evt.Cid
		}
		// update the expected document heads of peer collection subs
		for id := range s.nodePeerCollections[collectionID] {
			s.expectedDocHeads[id][evt.DocID] = evt.Cid
		}
	}
}

// waitForMergeEvents waits for all expected document heads to be merged to all nodes.
//
// Will fail the test if an event is not received within the expected time interval to prevent tests
// from running forever.
func waitForMergeEvents(
	s *state,
	action WaitForSync,
) {
	var timeout time.Duration
	if action.ExpectedTimeout != 0 {
		timeout = action.ExpectedTimeout
	} else {
		timeout = subscriptionTimeout * 10
	}

	ctx, cancel := context.WithTimeout(s.ctx, timeout)
	defer cancel()

	for nodeID, expect := range s.expectedDocHeads {
		// remove any docs that are already merged
		// up to the expected document head
		for key, val := range s.actualDocHeads[nodeID] {
			if head, ok := expect[key]; ok && head.String() == val.String() {
				delete(expect, key)
			}
		}
		// wait for all expected doc heads to be merged
		//
		// the order of merges does not matter as we only
		// expect the latest head to eventually be merged
		//
		// unexpected merge events are ignored
		for len(expect) > 0 {
			var evt event.Merge
			select {
			case msg := <-s.nodeEvents[nodeID].merge.Message():
				evt = msg.Data.(event.Merge)

			case <-ctx.Done():
				require.Fail(s.t, "timeout waiting for merge complete event")
			}

			head, ok := expect[evt.DocID]
			if ok && head.String() == evt.Cid.String() {
				delete(expect, evt.DocID)
			}
		}
	}
}
