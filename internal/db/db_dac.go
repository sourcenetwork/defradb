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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/dac"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/utils"
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
	opts ...options.Enumerable[options.AddDACPolicyOptions],
) (client.AddPolicyResult, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeAddDACPolicyPerm); err != nil {
		return client.AddPolicyResult{}, err
	}

	if !db.documentACP.HasValue() {
		return client.AddPolicyResult{}, client.ErrACPOperationButACPNotAvailable
	}

	policyID, err := db.documentACP.Value().AddPolicy(
		ctx,
		opt.Identity.Value(),
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
	opts ...options.Enumerable[options.AddDACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeAddDACRelationPerm); err != nil {
		return client.AddActorRelationshipResult{}, err
	}

	if !db.documentACP.HasValue() {
		return client.AddActorRelationshipResult{}, client.ErrACPOperationButACPNotAvailable
	}

	colOpt := options.WithIdentity(options.GetCollectionByName(), opt.Identity)
	collection, err := db.GetCollectionByName(ctx, collectionName, colOpt)
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
		opt.Identity.Value(),
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
	opts ...options.Enumerable[options.DeleteDACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeDeleteDACRelationPerm); err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}

	if !db.documentACP.HasValue() {
		return client.DeleteActorRelationshipResult{}, client.ErrACPOperationButACPNotAvailable
	}

	colOpt := options.WithIdentity(options.GetCollectionByName(), opt.Identity)
	collection, err := db.GetCollectionByName(ctx, collectionName, colOpt)
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
		opt.Identity.Value(),
		targetActor,
	)

	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}

	return client.DeleteActorRelationshipResult{RecordFound: recordFound}, nil
}
