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
	"encoding/json"
	"slices"
	"sync"
	"time"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/encryption"
)

// eventTimeout is the amount of time to wait
// for an event before timing out
const eventTimeout = 1 * time.Second

// waitForNetworkSetupEvents waits for p2p topic completed and
// replicator completed events to be published on the local node event bus.
func waitForNetworkSetupEvents(s *state, nodeID int) {
	cols, err := s.nodes[nodeID].GetAllP2PCollections(s.ctx)
	require.NoError(s.t, err)

	reps, err := s.nodes[nodeID].GetAllReplicators(s.ctx)
	require.NoError(s.t, err)

	replicatorEvents := len(reps)
	p2pTopicEvent := len(cols) > 0

	for p2pTopicEvent && replicatorEvents > 0 {
		select {
		case _, ok := <-s.nodeEvents[nodeID].replicator.Message():
			if !ok {
				require.Fail(s.t, "subscription closed waiting for network setup events")
			}
			replicatorEvents--

		case _, ok := <-s.nodeEvents[nodeID].p2pTopic.Message():
			if !ok {
				require.Fail(s.t, "subscription closed waiting for network setup events")
			}
			p2pTopicEvent = false

		case <-time.After(eventTimeout):
			s.t.Fatalf("timeout waiting for network setup events")
		}
	}
}

// waitForReplicatorConfigureEvent waits for a  node to publish a
// replicator completed event on the local event bus.
//
// Expected document heads will be updated for the targeted node.
func waitForReplicatorConfigureEvent(s *state, cfg ConfigureReplicator) {
	select {
	case _, ok := <-s.nodeEvents[cfg.SourceNodeID].replicator.Message():
		if !ok {
			require.Fail(s.t, "subscription closed waiting for replicator event")
		}

	case <-time.After(eventTimeout):
		require.Fail(s.t, "timeout waiting for replicator event")
	}

	// all previous documents should be merged on the subscriber node
	for key, val := range s.nodeP2P[cfg.SourceNodeID].actualDocHeads {
		s.nodeP2P[cfg.TargetNodeID].expectedDocHeads[key] = val
	}

	// update node connections and replicators
	s.nodeP2P[cfg.TargetNodeID].connections[cfg.SourceNodeID] = struct{}{}
	s.nodeP2P[cfg.SourceNodeID].connections[cfg.TargetNodeID] = struct{}{}
	s.nodeP2P[cfg.SourceNodeID].replicators[cfg.TargetNodeID] = struct{}{}
}

// waitForReplicatorConfigureEvent waits for a node to publish a
// replicator completed event on the local event bus.
func waitForReplicatorDeleteEvent(s *state, cfg DeleteReplicator) {
	select {
	case _, ok := <-s.nodeEvents[cfg.SourceNodeID].replicator.Message():
		if !ok {
			require.Fail(s.t, "subscription closed waiting for replicator event")
		}

	case <-time.After(eventTimeout):
		require.Fail(s.t, "timeout waiting for replicator event")
	}

	delete(s.nodeP2P[cfg.TargetNodeID].connections, cfg.SourceNodeID)
	delete(s.nodeP2P[cfg.SourceNodeID].connections, cfg.TargetNodeID)
	delete(s.nodeP2P[cfg.SourceNodeID].replicators, cfg.TargetNodeID)
}

// waitForSubscribeToCollectionEvent waits for a node to publish a
// p2p topic completed event on the local event bus.
//
// Expected document heads will be updated for the subscriber node.
func waitForSubscribeToCollectionEvent(s *state, action SubscribeToCollection) {
	select {
	case _, ok := <-s.nodeEvents[action.NodeID].p2pTopic.Message():
		if !ok {
			require.Fail(s.t, "subscription closed waiting for p2p topic event")
		}

	case <-time.After(eventTimeout):
		require.Fail(s.t, "timeout waiting for p2p topic event")
	}

	// update peer collections of target node
	for _, collectionIndex := range action.CollectionIDs {
		if collectionIndex == NonExistentCollectionID {
			continue // don't track non existent collections
		}
		s.nodeP2P[action.NodeID].peerCollections[collectionIndex] = struct{}{}
	}
}

// waitForSubscribeToCollectionEvent waits for a node to publish a
// p2p topic completed event on the local event bus.
func waitForUnsubscribeToCollectionEvent(s *state, action UnsubscribeToCollection) {
	select {
	case _, ok := <-s.nodeEvents[action.NodeID].p2pTopic.Message():
		if !ok {
			require.Fail(s.t, "subscription closed waiting for p2p topic event")
		}

	case <-time.After(eventTimeout):
		require.Fail(s.t, "timeout waiting for p2p topic event")
	}

	for _, collectionIndex := range action.CollectionIDs {
		if collectionIndex == NonExistentCollectionID {
			continue // don't track non existent collections
		}
		delete(s.nodeP2P[action.NodeID].peerCollections, collectionIndex)
	}
}

// waitForUpdateEvents waits for all selected nodes to publish an
// update event to the local event bus.
//
// Expected document heads will be updated for any connected nodes.
func waitForUpdateEvents(
	s *state,
	nodeID immutable.Option[int],
	docIDs map[string]struct{},
) {
	for i := 0; i < len(s.nodes); i++ {
		if nodeID.HasValue() && nodeID.Value() != i {
			continue // node is not selected
		}

		expect := make(map[string]struct{}, len(docIDs))
		for k := range docIDs {
			expect[k] = struct{}{}
		}

		for len(expect) > 0 {
			var evt event.Update
			select {
			case msg, ok := <-s.nodeEvents[i].update.Message():
				if !ok {
					require.Fail(s.t, "subscription closed waiting for update event")
				}
				evt = msg.Data.(event.Update)

			case <-time.After(eventTimeout):
				require.Fail(s.t, "timeout waiting for update event")
			}

			// make sure the event is expected
			_, ok := expect[evt.DocID]
			require.True(s.t, ok, "unexpected document update")
			delete(expect, evt.DocID)

			// we only need to update the network state if the nodes
			// are configured for networking
			if i < len(s.nodeConfigs) {
				updateNetworkState(s, i, evt)
			}
		}
	}
}

// waitForMergeEvents waits for all expected document heads to be merged to all nodes.
//
// Will fail the test if an event is not received within the expected time interval to prevent tests
// from running forever.
func waitForMergeEvents(s *state) {
	for nodeID := 0; nodeID < len(s.nodes); nodeID++ {
		expect := s.nodeP2P[nodeID].expectedDocHeads

		// remove any docs that are already merged
		// up to the expected document head
		for key, val := range s.nodeP2P[nodeID].actualDocHeads {
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
			case msg, ok := <-s.nodeEvents[nodeID].merge.Message():
				if !ok {
					require.Fail(s.t, "subscription closed waiting for merge complete event")
				}
				evt = msg.Data.(event.Merge)

			case <-time.After(30 * eventTimeout):
				require.Fail(s.t, "timeout waiting for merge complete event")
			}

			head, ok := expect[evt.DocID]
			if ok && head.String() == evt.Cid.String() {
				delete(expect, evt.DocID)
			}
		}
	}
}

// updateNetworkState updates the network state by checking which
// nodes should receive the updated document in the given update event.
func updateNetworkState(s *state, nodeID int, evt event.Update) {
	// find the correct collection index for this update
	collectionID := -1
	for i, c := range s.collections[nodeID] {
		if c.SchemaRoot() == evt.SchemaRoot {
			collectionID = i
		}
	}

	// update the actual document head on the node that updated it
	s.nodeP2P[nodeID].actualDocHeads[evt.DocID] = evt.Cid

	// update the expected document heads of replicator targets
	for id := range s.nodeP2P[nodeID].replicators {
		// replicator target nodes push updates to source nodes
		s.nodeP2P[id].expectedDocHeads[evt.DocID] = evt.Cid
	}

	// update the expected document heads of connected nodes
	for id := range s.nodeP2P[nodeID].connections {
		// connected nodes share updates of documents they have in common
		if _, ok := s.nodeP2P[id].actualDocHeads[evt.DocID]; ok {
			s.nodeP2P[id].expectedDocHeads[evt.DocID] = evt.Cid
		}
		// peer collection subscribers receive updates from any other subscriber node
		if _, ok := s.nodeP2P[id].peerCollections[collectionID]; ok {
			s.nodeP2P[id].expectedDocHeads[evt.DocID] = evt.Cid
		}
	}

	// make sure the event is published on the network before proceeding
	// this prevents nodes from missing messages that are sent before
	// subscriptions are setup
	time.Sleep(100 * time.Millisecond)
}

// getEventsForUpdateDoc returns a map of docIDs that should be
// published to the local event bus after an UpdateDoc action.
//
// This will take into account any primary documents that are patched as a result
// of the create or update.
func getEventsForUpdateDoc(s *state, action UpdateDoc) map[string]struct{} {
	var collection client.Collection
	if action.NodeID.HasValue() {
		collection = s.collections[action.NodeID.Value()][action.CollectionID]
	} else {
		collection = s.collections[0][action.CollectionID]
	}

	docID := s.docIDs[action.CollectionID][action.DocID]
	def := collection.Definition()

	docMap := make(map[string]any)
	err := json.Unmarshal([]byte(action.Doc), &docMap)
	require.NoError(s.t, err)

	expect := make(map[string]struct{})
	expect[docID.String()] = struct{}{}

	// check for any secondary relation fields that could publish an event
	for name, value := range docMap {
		field, ok := def.GetFieldByName(name)
		if !ok {
			continue // ignore unknown field
		}
		_, ok = field.GetSecondaryRelationField(def)
		if ok {
			expect[value.(string)] = struct{}{}
		}
	}

	return expect
}

// getEventsForCreateDoc returns a map of docIDs that should be
// published to the local event bus after a CreateDoc action.
//
// This will take into account any primary documents that are patched as a result
// of the create or update.
func getEventsForCreateDoc(s *state, action CreateDoc) map[string]struct{} {
	var collection client.Collection
	if action.NodeID.HasValue() {
		collection = s.collections[action.NodeID.Value()][action.CollectionID]
	} else {
		collection = s.collections[0][action.CollectionID]
	}

	docs, err := parseCreateDocs(action, collection)
	require.NoError(s.t, err)

	def := collection.Definition()
	expect := make(map[string]struct{})

	for _, doc := range docs {
		expect[doc.ID().String()] = struct{}{}

		// check for any secondary relation fields that could publish an event
		for f, v := range doc.Values() {
			field, ok := def.GetFieldByName(f.Name())
			if !ok {
				continue // ignore unknown field
			}
			_, ok = field.GetSecondaryRelationField(def)
			if ok {
				expect[v.Value().(string)] = struct{}{}
			}
		}
	}

	return expect
}

// waitForKeyRetrievedEvent waits for nodes to receive a key retrieved event.
// If targetNodeIDs is empty, all nodes will be waiting on the event sync.
// Otherwise, only the nodes with the specified IDs will be checked.
func waitForKeyRetrievedEvent(s *state, targetNodeIDs []int) {
	for nodeID := 0; nodeID < len(s.nodes); nodeID++ {
		// if we have target nodes and the current node is not in the list, skip it
		if len(targetNodeIDs) > 0 && !slices.Contains(targetNodeIDs, nodeID) {
			continue
		}

		select {
		case _, ok := <-s.nodeEvents[nodeID].encKeysRetrieved.Message():
			if !ok {
				require.Fail(s.t, "subscription closed waiting for key retrieved event")
			}

		case <-time.After(eventTimeout):
			require.Fail(s.t, "timeout waiting for key retrieved event")
		}
	}
}

func waitForSync(s *state, action WaitForSync) {
	count := int(action.Count)
	if count == 0 {
		count = 1
	}

	syncFunc := func(s *state, action WaitForSync) {
		if !action.Event.HasValue() || action.Event.Value() == event.MergeCompleteName {
			waitForMergeEvents(s)
		} else if action.Event.Value() == encryption.KeysRetrievedEventName {
			waitForKeyRetrievedEvent(s, action.NodeIDs)
		} else {
			require.Fail(s.t, "unsupported event type: %s", action.Event.Value())
		}
	}

	if count == 1 {
		syncFunc(s, action)
	} else {
		var wg sync.WaitGroup
		wg.Add(count)
		for i := 0; i < count; i++ {
			go func(s *state, action WaitForSync) {
				defer wg.Done()
				syncFunc(s, action)
			}(s, action)
		}
		wg.Wait()
	}
}

// getEventsForUpdateWithFilter returns a map of docIDs that should be
// published to the local event bus after a UpdateWithFilter action.
//
// This will take into account any primary documents that are patched as a result
// of the create or update.
func getEventsForUpdateWithFilter(
	s *state,
	action UpdateWithFilter,
	result *client.UpdateResult,
) map[string]struct{} {
	var collection client.Collection
	if action.NodeID.HasValue() {
		collection = s.collections[action.NodeID.Value()][action.CollectionID]
	} else {
		collection = s.collections[0][action.CollectionID]
	}

	var docPatch map[string]any
	err := json.Unmarshal([]byte(action.Updater), &docPatch)
	require.NoError(s.t, err)

	def := collection.Definition()
	expect := make(map[string]struct{})

	for _, docID := range result.DocIDs {
		expect[docID] = struct{}{}

		// check for any secondary relation fields that could publish an event
		for name, value := range docPatch {
			field, ok := def.GetFieldByName(name)
			if !ok {
				continue // ignore unknown field
			}
			_, ok = field.GetSecondaryRelationField(def)
			if ok {
				expect[value.(string)] = struct{}{}
			}
		}
	}

	return expect
}
