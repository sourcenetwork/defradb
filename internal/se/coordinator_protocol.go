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
	"strings"

	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/internal/datastore"
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

func (rc *Coordinator) processPushSEArtifactsRequest(
	ctx context.Context,
	req *PushSEArtifactsRequest,
) error {
	sb := strings.Builder{}
	for i, netArtifact := range req.Artifacts {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(netArtifact.DocID)
	}
	log.InfoContext(ctx, "Handle push SE artifacts",
		corelog.String("DocIDs", sb.String()), corelog.String("Sender", req.SenderID))

	artifacts := make([]secore.Artifact, len(req.Artifacts))
	for i, netArtifact := range req.Artifacts {
		artifacts[i] = secore.Artifact{
			DocID:        netArtifact.DocID,
			IndexID:      netArtifact.IndexID,
			SearchTag:    netArtifact.SearchTag,
			CollectionID: req.CollectionID,
		}
	}

	if err := StoreArtifacts(ctx, datastore.DatastoreFrom(rc.db.Rootstore()), artifacts); err != nil {
		return err
	}

	return nil
}

func (rc *Coordinator) processQuerySEArtifactsRequest(
	ctx context.Context,
	req *QuerySEArtifactsRequest,
) (QuerySEArtifactsReply, error) {
	matchingDocIDs, err := rc.querySEArtifactsFromDatastore(ctx, req)
	if err != nil {
		log.ErrorContextE(ctx, "Failed to query SE artifacts from datastore", err)
		return QuerySEArtifactsReply{}, err
	}

	log.InfoContext(ctx, "Handle SE artifacts query", corelog.String("DocIDs", strings.Join(matchingDocIDs, ", ")),
		corelog.String("Sender", req.SenderID))

	return QuerySEArtifactsReply{
		DocIDs: matchingDocIDs,
	}, nil
}

// querySEArtifactsFromDatastore queries SE artifacts from the local datastore
func (rc *Coordinator) querySEArtifactsFromDatastore(
	ctx context.Context,
	req *QuerySEArtifactsRequest,
) ([]string, error) {
	queries := make([]FieldQuery, len(req.Queries))
	for i, q := range req.Queries {
		queries[i] = FieldQuery(q)
	}

	return FetchDocIDs(ctx, datastore.DatastoreFrom(rc.db.Rootstore()), req.CollectionID, queries)
}
