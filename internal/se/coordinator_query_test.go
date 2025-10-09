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
)

func TestQuerySEArtifacts_WhenReplicatorsExist_ShouldQueryAndReturnDocIDs(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	queries := []fieldQuery{
		{
			FieldName: "field1",
			IndexID:   "index-1",
			SearchTag: []byte("tag-1"),
		},
	}

	setup.mockGetReplicatorsIDs([]string{setup.peerID})

	expectedReply := QuerySEArtifactsReply{DocIDs: []string{"doc-1", "doc-2"}}
	setup.mockQueryProto.EXPECT().SendRequest(mock.Anything, mock.Anything, setup.peerID).Return(expectedReply, nil)

	docIDs, err := setup.coordinator.QuerySEArtifacts(context.Background(), setup.collectionID, queries)

	require.NoError(t, err)
	require.Len(t, docIDs, 2)
	require.Contains(t, docIDs, "doc-1")
	require.Contains(t, docIDs, "doc-2")
}

func TestQuerySEArtifacts_WhenNoReplicators_ShouldReturnEmpty(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	queries := []fieldQuery{
		{
			FieldName: "field1",
			IndexID:   "index-1",
			SearchTag: []byte("tag-1"),
		},
	}

	setup.mockGetReplicatorsIDs([]string{})

	docIDs, err := setup.coordinator.QuerySEArtifacts(context.Background(), setup.collectionID, queries)

	require.NoError(t, err)
	require.Empty(t, docIDs)
}

func TestQuerySEArtifacts_WhenFirstReplicatorFails_ShouldTryNext(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	queries := []fieldQuery{
		{
			FieldName: "field1",
			IndexID:   "index-1",
			SearchTag: []byte("tag-1"),
		},
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

	docIDs, err := setup.coordinator.QuerySEArtifacts(context.Background(), setup.collectionID, queries)

	require.NoError(t, err)
	require.Len(t, docIDs, 2)
	require.Contains(t, docIDs, "doc-3")
	require.Contains(t, docIDs, "doc-4")
}

func TestQuerySEArtifacts_WhenAllReplicatorsFail_ShouldReturnError(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	queries := []fieldQuery{
		{
			FieldName: "field1",
			IndexID:   "index-1",
			SearchTag: []byte("tag-1"),
		},
	}

	peerID1 := "peer-1"
	peerID2 := "peer-2"
	setup.mockGetReplicatorsIDs([]string{peerID1, peerID2})

	setup.mockQueryProto.EXPECT().SendRequest(mock.Anything, mock.Anything, peerID1).
		Return(QuerySEArtifactsReply{}, fmt.Errorf("network error 1")).Once()

	setup.mockQueryProto.EXPECT().SendRequest(mock.Anything, mock.Anything, peerID2).
		Return(QuerySEArtifactsReply{}, fmt.Errorf("network error 2")).Once()

	docIDs, err := setup.coordinator.QuerySEArtifacts(context.Background(), setup.collectionID, queries)

	require.Error(t, err)
	require.Empty(t, docIDs)
	require.Contains(t, err.Error(), "network error 2")
}

func TestQuerySEArtifacts_WhenMultipleQueries_ShouldPassAllToReplicator(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	queries := []fieldQuery{
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

	docIDs, err := setup.coordinator.QuerySEArtifacts(context.Background(), setup.collectionID, queries)

	require.NoError(t, err)
	require.Len(t, docIDs, 3)
	require.Contains(t, docIDs, "doc-1")
	require.Contains(t, docIDs, "doc-2")
	require.Contains(t, docIDs, "doc-3")
}
