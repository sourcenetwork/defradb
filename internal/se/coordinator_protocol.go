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

	"github.com/sourcenetwork/defradb/event"
	secore "github.com/sourcenetwork/defradb/internal/se/core"
)

// seStoreProcessor implements CommProcessor for SE artifact storage
type seStoreProcessor struct {
	coordinator *Coordinator
}

func (proc *seStoreProcessor) ProcessRequest(
	ctx context.Context,
	req PushSEArtifactsRequest,
) (PushSEArtifactsReply, error) {
	return PushSEArtifactsReply{}, proc.coordinator.processPushSEArtifactsRequest(ctx, &req)
}

// seQueryProcessor implements CommProcessor for SE artifact queries
type seQueryProcessor struct {
	coordinator *Coordinator
}

func (proc *seQueryProcessor) ProcessRequest(
	ctx context.Context,
	req QuerySEArtifactsRequest,
) (QuerySEArtifactsReply, error) {
	return proc.coordinator.processQuerySEArtifactsRequest(ctx, &req)
}

func (coordinator *Coordinator) processPushSEArtifactsRequest(
	ctx context.Context,
	req *PushSEArtifactsRequest,
) error {
	artifacts := make([]secore.Artifact, len(req.Artifacts))
	for i, netArtifact := range req.Artifacts {
		artifacts[i] = secore.Artifact{
			DocID:        netArtifact.DocID,
			IndexID:      netArtifact.IndexID,
			SearchTag:    netArtifact.SearchTag,
			CollectionID: req.CollectionID,
		}
	}

	err := storeArtifacts(ctx, coordinator.db.Multistore(), artifacts)
	if err != nil {
		return err
	}

	// Group artifacts by docID to publish events
	docFieldsMap := make(map[string]map[string]struct{})
	for _, artifact := range artifacts {
		if _, exists := docFieldsMap[artifact.DocID]; !exists {
			docFieldsMap[artifact.DocID] = make(map[string]struct{})
		}
		docFieldsMap[artifact.DocID][artifact.IndexID] = struct{}{}
	}

	// Publish SEArtifactSyncComplete event for each document
	for docID, fieldsSet := range docFieldsMap {
		fieldNames := make([]string, 0, len(fieldsSet))
		for fieldName := range fieldsSet {
			fieldNames = append(fieldNames, fieldName)
		}

		coordinator.db.Events().Publish(event.NewMessage(event.SEArtifactReceivedName, event.SEArtifactReceived{
			DocID:        docID,
			CollectionID: req.CollectionID,
			FieldNames:   fieldNames,
		}))
	}

	return nil
}

func (coordinator *Coordinator) processQuerySEArtifactsRequest(
	ctx context.Context,
	req *QuerySEArtifactsRequest,
) (QuerySEArtifactsReply, error) {
	matchingDocIDs, err := coordinator.querySEArtifactsFromDatastore(ctx, req)
	if err != nil {
		return QuerySEArtifactsReply{}, err
	}

	return QuerySEArtifactsReply{
		DocIDs: matchingDocIDs,
	}, nil
}

// querySEArtifactsFromDatastore queries SE artifacts from the local datastore
func (coordinator *Coordinator) querySEArtifactsFromDatastore(
	ctx context.Context,
	req *QuerySEArtifactsRequest,
) ([]string, error) {
	queries := make([]fieldQuery, len(req.Queries))
	for i, q := range req.Queries {
		queries[i] = fieldQuery(q)
	}
	return fetchDocIDs(ctx, coordinator.db.Multistore(), req.CollectionID, queries)
}
