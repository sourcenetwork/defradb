// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package se_test

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
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/se"
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
		Return(se.PushSEArtifactsReply{}, fmt.Errorf("network error"))

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

			var retryInfo se.SERetryInfo
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




