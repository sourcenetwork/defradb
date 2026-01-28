// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"

	"github.com/sourcenetwork/defradb/acp/dac"
	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/immutable"
)

func (db *DB) DocumentACP() immutable.Option[dac.DocumentACP] {
	return db.documentACP
}

// PurgeDACState purges all document ACP state, and calls [Close()] on the acp instance before returning.
//
// This will close the acp system, reset it's state (purge then restart), and finally close it.
//
// Note: all document ACP state will be lost, and won't be recoverable.
func (db *DB) PurgeDACState(ctx context.Context) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	// Purge document acp state and keep it closed.
	if db.documentACP.HasValue() {
		documentACP := db.documentACP.Value()
		err := documentACP.ResetState(ctx)
		if err != nil {
			// for now we will just log this error, since SourceHub ACP doesn't yet
			// implement the ResetState.
			log.ErrorE("Failed to reset document ACP state", err)
		}
	}

	return nil
}

func (db *DB) AddDACPolicy(
	ctx context.Context,
	policy string,
) (client.AddPolicyResult, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if err := db.checkNodeAccess(ctx, acpTypes.NodeDACPolicyAddPerm); err != nil {
		return client.AddPolicyResult{}, err
	}

	if !db.documentACP.HasValue() {
		return client.AddPolicyResult{}, client.ErrACPOperationButACPNotAvailable
	}

	policyID, err := db.documentACP.Value().AddPolicy(
		ctx,
		identity.FromContext(ctx).Value(),
		policy,
	)
	if err != nil {
		return client.AddPolicyResult{}, err
	}

	return client.AddPolicyResult{PolicyID: policyID}, nil
}

func (db *DB) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if err := db.checkNodeAccess(ctx, acpTypes.NodeDACRelationAddPerm); err != nil {
		return client.AddActorRelationshipResult{}, err
	}

	if !db.documentACP.HasValue() {
		return client.AddActorRelationshipResult{}, client.ErrACPOperationButACPNotAvailable
	}

	collection, err := db.GetCollectionByName(ctx, collectionName)
	if err != nil {
		return client.AddActorRelationshipResult{}, err
	}

	policyID, resourceName, hasPolicy := acpDB.IsPermissioned(collection)
	if !hasPolicy {
		return client.AddActorRelationshipResult{}, client.ErrACPOperationButCollectionHasNoPolicy
	}

	exists, err := db.documentACP.Value().AddDocActorRelationship(
		ctx,
		policyID,
		resourceName,
		docID,
		relation,
		identity.FromContext(ctx).Value(),
		targetActor,
	)

	if err != nil {
		return client.AddActorRelationshipResult{}, err
	}

	if !exists {
		err = db.publishDocUpdateEvent(ctx, docID, collection)
		if err != nil {
			return client.AddActorRelationshipResult{}, err
		}
	}

	return client.AddActorRelationshipResult{ExistedAlready: exists}, nil
}

func (db *DB) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if err := db.checkNodeAccess(ctx, acpTypes.NodeDACRelationDeletePerm); err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}

	if !db.documentACP.HasValue() {
		return client.DeleteActorRelationshipResult{}, client.ErrACPOperationButACPNotAvailable
	}

	collection, err := db.GetCollectionByName(ctx, collectionName)
	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}

	policyID, resourceName, hasPolicy := acpDB.IsPermissioned(collection)
	if !hasPolicy {
		return client.DeleteActorRelationshipResult{}, client.ErrACPOperationButCollectionHasNoPolicy
	}

	recordFound, err := db.documentACP.Value().DeleteDocActorRelationship(
		ctx,
		policyID,
		resourceName,
		docID,
		relation,
		identity.FromContext(ctx).Value(),
		targetActor,
	)

	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}

	return client.DeleteActorRelationshipResult{RecordFound: recordFound}, nil
}
