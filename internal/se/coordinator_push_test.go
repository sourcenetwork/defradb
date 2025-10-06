// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package se

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/datastore"
	message "github.com/sourcenetwork/defradb/internal/db/p2p/message"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func TestReplicationCoordinator_WhenHandlePushToReplicatorsCalled_ShouldPushSEArtifactsToPeers(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	requestChan := setup.expectSEArtifactPush()

	evt := setup.makeUpdateEvent()
	err := setup.coordinator.HandlePushToReplicators(context.Background(), evt)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		select {
		case req := <-requestChan:
			return req.CollectionID == setup.collectionID && len(req.Artifacts) > 0
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond, "SE artifacts should be pushed to replicator with expected data")
}

func TestReplicationCoordinator_WhenBlockFailsToDeserialize_ShouldReturnError(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	updateEvent := event.Update{
		DocID:        setup.docID,
		CollectionID: setup.collectionID,
		Block:        []byte("invalid-block-data"),
	}
	err := setup.coordinator.HandlePushToReplicators(context.Background(), updateEvent)
	require.Error(t, err, "Should return error when block fails to deserialize")

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenNonCompositeBlock_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	updateEvent := event.Update{
		DocID:        setup.docID,
		CollectionID: setup.collectionID,
		Block:        setup.createNonCompositeBlock(),
	}
	err := setup.coordinator.HandlePushToReplicators(context.Background(), updateEvent)
	require.NoError(t, err)

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenGetCollectionsFails_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	setup.mockDB.EXPECT().GetCollections(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("database error"))

	evt := setup.makeUpdateEvent()
	err := setup.coordinator.HandlePushToReplicators(context.Background(), evt)
	require.Error(t, err) // Should return error when GetCollections fails

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenCollectionNotFound_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	setup.mockGetCollections()

	evt := setup.makeUpdateEvent()
	err := setup.coordinator.HandlePushToReplicators(context.Background(), evt)
	require.Error(t, err, "Should return error when collection not found")

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenInvalidDocID_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	setup.docID = "invalid-doc-id"
	setup.mockGetCollections(setup.createMockCollectionWithDocument())

	evt := setup.makeUpdateEvent()
	err := setup.coordinator.HandlePushToReplicators(context.Background(), evt)
	require.Error(t, err, "Should return error when doc ID is invalid")

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenDocumentNotFound_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	mockCollection := setup.createMockCollection()
	mockCollection.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, client.ErrDocumentNotFoundOrNotAuthorized).Maybe()
	setup.mockGetCollections(mockCollection)

	setup.mockGetReplicatorsIDs([]string{})

	evt := setup.makeUpdateEvent()
	err := setup.coordinator.HandlePushToReplicators(context.Background(), evt)
	require.NoError(t, err)

	require.Empty(setup.t, setup.mockStorageProto.Calls, "No SE artifacts should be pushed")
}

func TestReplicationCoordinator_WhenDocumentGetFails_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	mockCollection := setup.createMockCollection()
	mockCollection.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("database error")).Maybe()
	setup.mockGetCollections(mockCollection)

	evt := setup.makeUpdateEvent()
	err := setup.coordinator.HandlePushToReplicators(context.Background(), evt)
	require.Error(t, err, "Should return error when document get fails")

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenNoEncryptedIndexes_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	mockCollection := setup.createMockCollection()
	mockCollection.EXPECT().ListEncryptedIndexes(mock.Anything).Return(
		[]client.EncryptedIndexDescription{}, nil).Maybe()

	ver := setup.createCollectionVersion()
	ver.EncryptedIndexes = []client.EncryptedIndexDescription{}
	mockCollection.EXPECT().Version().Return(ver).Maybe()

	doc, err := client.NewDocFromMap(map[string]any{"age": 21}, ver)
	mockCollection.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(doc, err).Maybe()

	setup.mockGetCollections(mockCollection)

	setup.mockGetReplicatorsIDs([]string{})

	evt := setup.makeUpdateEvent()
	err = setup.coordinator.HandlePushToReplicators(context.Background(), evt)
	require.NoError(t, err)

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenPushToReplicatorFails_ShouldStoreRetryInPeerstore(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	setup.mockGetCollections(setup.createMockCollectionWithDocument())

	setup.mockGetReplicatorsIDs([]string{setup.peerID})

	setup.mockStorageProto.EXPECT().SendRequest(mock.Anything, mock.Anything, setup.peerID).
		Return(PushSEArtifactsReply{}, fmt.Errorf("network error"))

	evt := setup.makeUpdateEvent()
	err := setup.coordinator.HandlePushToReplicators(context.Background(), evt)
	require.NoError(t, err) // Error is stored in retry, not returned

	require.Eventually(t, func() bool {
		ps := datastore.PeerstoreFrom(setup.rootstore)
		retryKey := keys.NewPeerstoreSERetry(setup.peerID, setup.collectionID, setup.docID)
		has, err := ps.Has(context.Background(), retryKey.Bytes())
		require.NoError(t, err)

		if has {
			value, err := ps.Get(context.Background(), retryKey.Bytes())
			require.NoError(t, err)

			var retryInfo SERetryInfo
			err = cbor.Unmarshal(value, &retryInfo)
			require.NoError(t, err)

			require.Equal(t, setup.docID, retryInfo.DocID)
			require.Equal(t, setup.collectionID, retryInfo.CollectionID)
			require.Contains(t, retryInfo.FieldNames, setup.fieldName)
			require.Equal(t, 0, retryInfo.NumRetries)
			require.False(t, retryInfo.Retrying)
			require.NotZero(t, retryInfo.NextRetry)

			return true
		}
		return false
	}, time.Second, 10*time.Millisecond, "Retry data should be stored in peerstore after push failure")
}

func TestProcessPushSEArtifactsRequest_WhenArtifactsProvided_ShouldStoreInDatastore(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	req := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-123",
		},
		CollectionID: setup.collectionID,
		Artifacts: []SEArtifact{
			{
				DocID:     setup.docID,
				IndexID:   "index-1",
				SearchTag: []byte("search-tag-1"),
			},
		},
	}

	err := setup.coordinator.processPushSEArtifactsRequest(context.Background(), req)
	require.NoError(t, err)

	ds := datastore.DatastoreFrom(setup.rootstore)
	key := keys.DatastoreSE{
		CollectionID: setup.collectionID,
		IndexID:      "index-1",
		SearchTag:    []byte("search-tag-1"),
		DocID:        setup.docID,
	}

	has, err := ds.Has(context.Background(), key.Bytes())
	require.NoError(t, err)
	require.True(t, has, "Artifact should be stored in datastore")
}

func TestProcessPushSEArtifactsRequest_WhenMultipleArtifacts_ShouldStoreAll(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	req := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-123",
		},
		CollectionID: setup.collectionID,
		Artifacts: []SEArtifact{
			{
				DocID:     setup.docID,
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
			{
				DocID:     "doc-2",
				IndexID:   "index-2",
				SearchTag: []byte("tag-2"),
			},
			{
				DocID:     "doc-3",
				IndexID:   "index-3",
				SearchTag: []byte("tag-3"),
			},
		},
	}

	err := setup.coordinator.processPushSEArtifactsRequest(context.Background(), req)
	require.NoError(t, err)

	ds := datastore.DatastoreFrom(setup.rootstore)

	for i, artifact := range req.Artifacts {
		key := keys.DatastoreSE{
			CollectionID: setup.collectionID,
			IndexID:      artifact.IndexID,
			SearchTag:    artifact.SearchTag,
			DocID:        artifact.DocID,
		}

		has, err := ds.Has(context.Background(), key.Bytes())
		require.NoError(t, err)
		require.True(t, has, "Artifact %d should be stored in datastore", i+1)
	}
}

func TestProcessPushSEArtifactsRequest_WhenEmptyArtifacts_ShouldSucceed(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	req := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-123",
		},
		CollectionID: setup.collectionID,
		Artifacts:    []SEArtifact{},
	}

	err := setup.coordinator.processPushSEArtifactsRequest(context.Background(), req)
	require.NoError(t, err, "Should succeed with empty artifacts list")
}

func TestProcessPushSEArtifactsRequest_WhenDuplicateArtifacts_ShouldOverwrite(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	artifact := SEArtifact{
		DocID:     setup.docID,
		IndexID:   "index-1",
		SearchTag: []byte("search-tag"),
	}

	req1 := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-1",
		},
		CollectionID: setup.collectionID,
		Artifacts:    []SEArtifact{artifact},
	}

	err := setup.coordinator.processPushSEArtifactsRequest(context.Background(), req1)
	require.NoError(t, err)

	req2 := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-2",
		},
		CollectionID: setup.collectionID,
		Artifacts:    []SEArtifact{artifact},
	}

	err = setup.coordinator.processPushSEArtifactsRequest(context.Background(), req2)
	require.NoError(t, err, "Should succeed when storing duplicate artifact")

	ds := datastore.DatastoreFrom(setup.rootstore)
	key := keys.DatastoreSE{
		CollectionID: setup.collectionID,
		IndexID:      artifact.IndexID,
		SearchTag:    artifact.SearchTag,
		DocID:        artifact.DocID,
	}

	has, err := ds.Has(context.Background(), key.Bytes())
	require.NoError(t, err)
	require.True(t, has, "Artifact should exist in datastore")
}
