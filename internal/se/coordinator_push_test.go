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

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
)

func TestReplicationCoordinator_WhenHandlePushToReplicatorsCalled_ShouldPushSEArtifactsToPeers(t *testing.T) {
	ctx := context.Background()
	setup := newTestSetup(t)
	defer setup.close()

	requestChan := setup.expectSEArtifactPush(ctx)

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
	ctx := context.Background()
	setup := newTestSetup(t)
	defer setup.close()

	setup.docID = "invalid-doc-id"
	setup.mockGetCollections(setup.createMockCollectionWithDocument(ctx))

	evt := setup.makeUpdateEvent()
	err := setup.coordinator.HandlePushToReplicators(context.Background(), evt)
	require.Error(t, err, "Should return error when doc ID is invalid")

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenDocumentNotFound_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	mockCollection := setup.createMockCollection()
	mockCollection.EXPECT().GetDocument(mock.Anything, mock.Anything, mock.Anything).
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
	mockCollection.EXPECT().GetDocument(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("database error")).Maybe()
	setup.mockGetCollections(mockCollection)

	evt := setup.makeUpdateEvent()
	err := setup.coordinator.HandlePushToReplicators(context.Background(), evt)
	require.Error(t, err, "Should return error when document get fails")

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenNoEncryptedIndexes_ShouldNotPushToPeer(t *testing.T) {
	ctx := context.Background()
	setup := newTestSetup(t)
	defer setup.close()

	mockCollection := setup.createMockCollection()
	mockCollection.EXPECT().ListEncryptedIndexes(mock.Anything).Return(
		[]client.EncryptedIndexDescription{}, nil).Maybe()

	ver := setup.createCollectionVersion()
	ver.EncryptedIndexes = []client.EncryptedIndexDescription{}
	mockCollection.EXPECT().Version().Return(ver).Maybe()

	doc, err := client.NewDocFromMap(ctx, map[string]any{"age": 21}, ver)
	mockCollection.EXPECT().GetDocument(mock.Anything, mock.Anything, mock.Anything).Return(doc, err).Maybe()

	setup.mockGetCollections(mockCollection)

	setup.mockGetReplicatorsIDs([]string{})

	evt := setup.makeUpdateEvent()
	err = setup.coordinator.HandlePushToReplicators(context.Background(), evt)
	require.NoError(t, err)

	setup.waitForNoCalls()
}
