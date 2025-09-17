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

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	message "github.com/sourcenetwork/defradb/internal/db/p2p/message"
)

func TestProcessQuerySEArtifactsRequest_WhenMatchingArtifacts_ShouldReturnDocIDs(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	artifacts := []SEArtifact{
		{
			DocID:     "doc-1",
			IndexID:   "index-1",
			SearchTag: []byte("tag-1"),
		},
		{
			DocID:     "doc-2",
			IndexID:   "index-1",
			SearchTag: []byte("tag-1"),
		},
		{
			DocID:     "doc-3",
			IndexID:   "index-2",
			SearchTag: []byte("tag-2"),
		},
	}

	storeReq := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-setup",
		},
		CollectionID: setup.collectionID,
		Artifacts:    artifacts,
	}
	err := setup.coordinator.processPushSEArtifactsRequest(context.Background(), storeReq)
	require.NoError(t, err)

	queryReq := &QuerySEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-query",
		},
		CollectionID: setup.collectionID,
		Queries: []SEFieldQuery{
			{
				FieldName: "field1",
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
		},
	}

	reply, err := setup.coordinator.processQuerySEArtifactsRequest(context.Background(), queryReq)
	require.NoError(t, err)
	require.NotNil(t, reply.DocIDs)
	require.Len(t, reply.DocIDs, 2, "Should return 2 matching documents")
	require.Contains(t, reply.DocIDs, "doc-1")
	require.Contains(t, reply.DocIDs, "doc-2")
}

func TestProcessQuerySEArtifactsRequest_WhenNoMatchingArtifacts_ShouldReturnEmpty(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	storeReq := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-setup",
		},
		CollectionID: setup.collectionID,
		Artifacts: []SEArtifact{
			{
				DocID:     "doc-1",
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
		},
	}
	err := setup.coordinator.processPushSEArtifactsRequest(context.Background(), storeReq)
	require.NoError(t, err)

	queryReq := &QuerySEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-query",
		},
		CollectionID: setup.collectionID,
		Queries: []SEFieldQuery{
			{
				FieldName: "field1",
				IndexID:   "non-existent-index",
				SearchTag: []byte("non-existent-tag"),
			},
		},
	}

	reply, err := setup.coordinator.processQuerySEArtifactsRequest(context.Background(), queryReq)
	require.NoError(t, err)
	require.Empty(t, reply.DocIDs, "Should return empty list when no matches found")
}

func TestProcessQuerySEArtifactsRequest_WhenMultipleQueries_ShouldReturnUnion(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	artifacts := []SEArtifact{
		{
			DocID:     "doc-1",
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
	}

	storeReq := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-setup",
		},
		CollectionID: setup.collectionID,
		Artifacts:    artifacts,
	}
	err := setup.coordinator.processPushSEArtifactsRequest(context.Background(), storeReq)
	require.NoError(t, err)

	queryReq := &QuerySEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-query",
		},
		CollectionID: setup.collectionID,
		Queries: []SEFieldQuery{
			{
				FieldName: "field1",
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
			{
				FieldName: "field2",
				IndexID:   "index-2",
				SearchTag: []byte("tag-2"),
			},
		},
	}

	reply, err := setup.coordinator.processQuerySEArtifactsRequest(context.Background(), queryReq)
	require.NoError(t, err)
	require.Len(t, reply.DocIDs, 2, "Should return union of matching documents")
	require.Contains(t, reply.DocIDs, "doc-1")
	require.Contains(t, reply.DocIDs, "doc-2")
}

func TestProcessQuerySEArtifactsRequest_WhenEmptyQueries_ShouldReturnEmpty(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	queryReq := &QuerySEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-query",
		},
		CollectionID: setup.collectionID,
		Queries:      []SEFieldQuery{},
	}

	reply, err := setup.coordinator.processQuerySEArtifactsRequest(context.Background(), queryReq)
	require.NoError(t, err)
	require.Empty(t, reply.DocIDs, "Should return empty list for empty queries")
}

func TestProcessQuerySEArtifactsRequest_WhenDifferentCollections_ShouldOnlyReturnFromSpecified(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	artifact1 := SEArtifact{
		DocID:     "doc-1",
		IndexID:   "index-1",
		SearchTag: []byte("tag-1"),
	}

	storeReq1 := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-setup",
		},
		CollectionID: setup.collectionID,
		Artifacts:    []SEArtifact{artifact1},
	}
	err := setup.coordinator.processPushSEArtifactsRequest(context.Background(), storeReq1)
	require.NoError(t, err)

	artifact2 := SEArtifact{
		DocID:     "doc-2",
		IndexID:   "index-1",
		SearchTag: []byte("tag-1"),
	}

	storeReq2 := &PushSEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-setup",
		},
		CollectionID: "other-collection",
		Artifacts:    []SEArtifact{artifact2},
	}
	err = setup.coordinator.processPushSEArtifactsRequest(context.Background(), storeReq2)
	require.NoError(t, err)

	queryReq := &QuerySEArtifactsRequest{
		MetaData: message.MetaData{
			SenderID: "peer-query",
		},
		CollectionID: setup.collectionID,
		Queries: []SEFieldQuery{
			{
				FieldName: "field1",
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
		},
	}

	reply, err := setup.coordinator.processQuerySEArtifactsRequest(context.Background(), queryReq)
	require.NoError(t, err)
	require.Len(t, reply.DocIDs, 1, "Should only return documents from specified collection")
	require.Contains(t, reply.DocIDs, "doc-1")
	require.NotContains(t, reply.DocIDs, "doc-2", "Should not return documents from other collections")
}

func TestHandleQuerySEArtifactsEvent_WhenReplicatorsExist_ShouldQueryAndReturnDocIDs(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	responseChan := make(chan SEArtifactsResult, 1)
	evt := RequestSEArtifactsEvent{
		CollectionID: setup.collectionID,
		Queries: []FieldQuery{
			{
				FieldName: "field1",
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
		},
		Response: responseChan,
	}

	setup.mockGetReplicatorsIDs([]string{setup.peerID})

	expectedReply := QuerySEArtifactsReply{DocIDs: []string{"doc-1", "doc-2"}}
	setup.mockQueryProto.EXPECT().SendRequest(mock.Anything, mock.Anything, setup.peerID).Return(expectedReply, nil)

	setup.coordinator.handleQuerySEArtifactsEvent(evt)

	result := <-responseChan
	require.NoError(t, result.Error)
	require.Len(t, result.DocIDs, 2)
	require.Contains(t, result.DocIDs, "doc-1")
	require.Contains(t, result.DocIDs, "doc-2")
}

func TestHandleQuerySEArtifactsEvent_WhenNoReplicators_ShouldReturnEmpty(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	responseChan := make(chan SEArtifactsResult, 1)
	evt := RequestSEArtifactsEvent{
		CollectionID: setup.collectionID,
		Queries: []FieldQuery{
			{
				FieldName: "field1",
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
		},
		Response: responseChan,
	}

	setup.mockGetReplicatorsIDs([]string{})

	setup.coordinator.handleQuerySEArtifactsEvent(evt)

	result := <-responseChan
	require.NoError(t, result.Error)
	require.Empty(t, result.DocIDs)
}

func TestHandleQuerySEArtifactsEvent_WhenFirstReplicatorFails_ShouldTryNext(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	responseChan := make(chan SEArtifactsResult, 1)
	evt := RequestSEArtifactsEvent{
		CollectionID: setup.collectionID,
		Queries: []FieldQuery{
			{
				FieldName: "field1",
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
		},
		Response: responseChan,
	}

	peerID1 := "peer-1"
	peerID2 := "peer-2"
	setup.mockGetReplicatorsIDs([]string{peerID1, peerID2})

	setup.mockQueryProto.EXPECT().SendRequest(mock.Anything, mock.Anything, peerID1).
		Return(QuerySEArtifactsReply{}, fmt.Errorf("network error")).Once()

	expectedReply := QuerySEArtifactsReply{
		DocIDs: []string{"doc-3", "doc-4"},
	}
	setup.mockQueryProto.EXPECT().SendRequest(mock.Anything, mock.Anything, peerID2).Return(expectedReply, nil).Once()

	setup.coordinator.handleQuerySEArtifactsEvent(evt)

	result := <-responseChan
	require.NoError(t, result.Error)
	require.Len(t, result.DocIDs, 2)
	require.Contains(t, result.DocIDs, "doc-3")
	require.Contains(t, result.DocIDs, "doc-4")
}

func TestHandleQuerySEArtifactsEvent_WhenAllReplicatorsFail_ShouldReturnError(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	responseChan := make(chan SEArtifactsResult, 1)
	evt := RequestSEArtifactsEvent{
		CollectionID: setup.collectionID,
		Queries: []FieldQuery{
			{
				FieldName: "field1",
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
		},
		Response: responseChan,
	}

	peerID1 := "peer-1"
	peerID2 := "peer-2"
	setup.mockGetReplicatorsIDs([]string{peerID1, peerID2})

	setup.mockQueryProto.EXPECT().SendRequest(mock.Anything, mock.Anything, peerID1).
		Return(QuerySEArtifactsReply{}, fmt.Errorf("network error 1")).Once()

	setup.mockQueryProto.EXPECT().SendRequest(mock.Anything, mock.Anything, peerID2).
		Return(QuerySEArtifactsReply{}, fmt.Errorf("network error 2")).Once()

	setup.coordinator.handleQuerySEArtifactsEvent(evt)

	result := <-responseChan
	require.Error(t, result.Error)
	require.Empty(t, result.DocIDs)
	require.Contains(t, result.Error.Error(), "network error 2")
}

func TestHandleQuerySEArtifactsEvent_WhenMultipleQueries_ShouldPassAllToReplicator(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	responseChan := make(chan SEArtifactsResult, 1)
	evt := RequestSEArtifactsEvent{
		CollectionID: setup.collectionID,
		Queries: []FieldQuery{
			{
				FieldName: "field1",
				IndexID:   "index-1",
				SearchTag: []byte("tag-1"),
			},
			{
				FieldName: "field2",
				IndexID:   "index-2",
				SearchTag: []byte("tag-2"),
			},
			{
				FieldName: "field3",
				IndexID:   "index-3",
				SearchTag: []byte("tag-3"),
			},
		},
		Response: responseChan,
	}

	setup.mockGetReplicatorsIDs([]string{setup.peerID})

	expectedReply := QuerySEArtifactsReply{
		DocIDs: []string{"doc-1", "doc-2", "doc-3"},
	}
	setup.mockQueryProto.EXPECT().SendRequest(
		mock.Anything,
		mock.MatchedBy(func(req QuerySEArtifactsRequest) bool {
			return req.CollectionID == setup.collectionID && len(req.Queries) == 3
		}),
		setup.peerID,
	).Return(expectedReply, nil)

	setup.coordinator.handleQuerySEArtifactsEvent(evt)

	result := <-responseChan
	require.NoError(t, result.Error)
	require.Len(t, result.DocIDs, 3)
	require.Contains(t, result.DocIDs, "doc-1")
	require.Contains(t, result.DocIDs, "doc-2")
	require.Contains(t, result.DocIDs, "doc-3")
}
