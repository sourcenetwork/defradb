// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package dac

import (
	"context"
	"strconv"

	protoTypes "github.com/cosmos/gogoproto/types"

	"github.com/sourcenetwork/corelog"
	"github.com/valyala/fastjson"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
)

var _ acp.ACPSystemClient = (*SourceHubDocumentACP)(nil)

var _ DocumentACP = (*bridgeDocumentACP)(nil)

// bridgeDocumentACP wraps an [ACPSystemClient], hosting the DefraDB specific logic away
// from ACP client specific code.
type bridgeDocumentACP struct {
	clientACP   acp.ACPSystemClient
	supportsP2P bool
}

func (a *bridgeDocumentACP) Init(ctx context.Context, path string) {
	a.clientACP.Init(ctx, path)
}

func (a *bridgeDocumentACP) Start(ctx context.Context) error {
	return a.clientACP.Start(ctx)
}

func (a *bridgeDocumentACP) AddPolicy(ctx context.Context, creator identity.Identity, policy string) (string, error) {
	// Having a creator identity is a MUST requirement for adding a policy.
	if creator.DID == "" {
		return "", acp.ErrPolicyCreatorMustNotBeEmpty
	}

	if policy == "" {
		return "", acp.ErrPolicyDataMustNotBeEmpty
	}

	marshalType := acpTypes.PolicyMarshalType_YAML
	if isJSON := fastjson.Validate(policy) == nil; isJSON { // Detect JSON format.
		marshalType = acpTypes.PolicyMarshalType_JSON
	}

	policyID, err := a.clientACP.AddPolicy(
		ctx,
		creator,
		policy,
		marshalType,
		protoTypes.TimestampNow(),
	)

	if err != nil {
		return "", acp.NewErrFailedToAddPolicyWithACP(err, "Local", creator.DID)
	}

	log.InfoContext(ctx, "Created Policy", corelog.Any("PolicyID", policyID))

	return policyID, nil
}

func (a *bridgeDocumentACP) ValidateResourceInterface(
	ctx context.Context,
	policyID string,
	resourceName string,
) error {
	var err error
	switch a.clientACP.(type) {
	case *LocalDocumentACP:
		err = acp.ValidateResourceInterfaceAccordingToACPSystem(
			ctx,
			policyID,
			resourceName,
			acpTypes.LocalDocumentACP,
			a.clientACP,
		)
	case *SourceHubDocumentACP:
		err = acp.ValidateResourceInterfaceAccordingToACPSystem(
			ctx,
			policyID,
			resourceName,
			acpTypes.SourceHubDocumentACP,
			a.clientACP,
		)
	default:
		return acp.ErrInvalidACPSystem
	}

	return err
}

func (a *bridgeDocumentACP) RegisterDocObject(
	ctx context.Context,
	identity identity.Identity,
	policyID string,
	resourceName string,
	docID string,
) error {
	err := a.clientACP.RegisterObject(
		ctx,
		identity,
		policyID,
		resourceName,
		docID,
		protoTypes.TimestampNow(),
	)

	if err != nil {
		return acp.NewErrFailedToRegisterDocWithACP(err, "Local", policyID, identity.DID, resourceName, docID)
	}

	return nil
}

func (a *bridgeDocumentACP) IsDocRegistered(
	ctx context.Context,
	policyID string,
	resourceName string,
	docID string,
) (bool, error) {
	maybeActor, err := a.clientACP.ObjectOwner(
		ctx,
		policyID,
		resourceName,
		docID,
	)
	if err != nil {
		return false, acp.NewErrFailedToCheckIfDocIsRegisteredWithACP(err, "Local", policyID, resourceName, docID)
	}

	return maybeActor.HasValue(), nil
}

func (a *bridgeDocumentACP) CheckDocAccess(
	ctx context.Context,
	permission acpTypes.DocumentResourcePermission,
	actorID string,
	policyID string,
	resourceName string,
	docID string,
) (bool, error) {
	// We grant "read" access even if the identity does not explicitly have the "read" permission,
	// as long as they have any of the permissions that imply read access.
	if permission == acpTypes.DocumentReadPerm {
		var canRead bool = false
		var withPermission string
		var err error

		for _, permissionThatImpliesRead := range acpTypes.ImplyDocumentReadPerm {
			canRead, err = a.clientACP.VerifyAccessRequest(
				ctx,
				permissionThatImpliesRead,
				actorID,
				policyID,
				resourceName,
				docID,
			)

			if err != nil {
				return false, acp.NewErrFailedToVerifyDocAccessWithACP(
					err,
					"Local",
					permissionThatImpliesRead.String(),
					policyID,
					actorID,
					resourceName,
					docID,
				)
			}

			if canRead {
				withPermission = permissionThatImpliesRead.String()
				break
			}
		}

		log.InfoContext(
			ctx,
			"Document readable="+strconv.FormatBool(canRead),
			corelog.Any("Permission", withPermission),
			corelog.Any("PolicyID", policyID),
			corelog.Any("Resource", resourceName),
			corelog.Any("ActorID", actorID),
			corelog.Any("DocID", docID),
		)

		return canRead, nil
	}

	hasAccess, err := a.clientACP.VerifyAccessRequest(
		ctx,
		permission,
		actorID,
		policyID,
		resourceName,
		docID,
	)

	if err != nil {
		return false, acp.NewErrFailedToVerifyDocAccessWithACP(
			err,
			"Local",
			permission.String(),
			policyID,
			actorID,
			resourceName,
			docID,
		)
	}

	log.InfoContext(
		ctx,
		"Document accessible="+strconv.FormatBool(hasAccess),
		corelog.Any("Permission", permission),
		corelog.Any("PolicyID", policyID),
		corelog.Any("Resource", resourceName),
		corelog.Any("ActorID", actorID),
		corelog.Any("DocID", docID),
	)

	return hasAccess, nil
}

func (a *bridgeDocumentACP) AddDocActorRelationship(
	ctx context.Context,
	policyID string,
	resourceName string,
	docID string,
	relation string,
	requestActor identity.Identity,
	targetActor string,
) (bool, error) {
	if policyID == "" ||
		resourceName == "" ||
		docID == "" ||
		relation == "" ||
		requestActor == (identity.Identity{}) ||
		targetActor == "" {
		return false, acp.NewErrMissingRequiredArgToAddDocActorRelationship(
			policyID,
			resourceName,
			docID,
			relation,
			requestActor.DID,
			targetActor,
		)
	}

	exists, err := a.clientACP.AddActorRelationship(
		ctx,
		policyID,
		resourceName,
		docID,
		relation,
		requestActor,
		targetActor,
		protoTypes.TimestampNow(),
	)

	if err != nil {
		return false, acp.NewErrFailedToAddDocActorRelationshipWithACP(
			err,
			"Local",
			policyID,
			resourceName,
			docID,
			relation,
			requestActor.DID,
			targetActor,
		)
	}

	log.InfoContext(
		ctx,
		"Document and actor relationship set",
		corelog.Any("PolicyID", policyID),
		corelog.Any("ResourceName", resourceName),
		corelog.Any("DocID", docID),
		corelog.Any("Relation", relation),
		corelog.Any("RequestActor", requestActor.DID),
		corelog.Any("TargetActor", targetActor),
		corelog.Any("Existed", exists),
	)

	return exists, nil
}

func (a *bridgeDocumentACP) DeleteDocActorRelationship(
	ctx context.Context,
	policyID string,
	resourceName string,
	docID string,
	relation string,
	requestActor identity.Identity,
	targetActor string,
) (bool, error) {
	if policyID == "" ||
		resourceName == "" ||
		docID == "" ||
		relation == "" ||
		requestActor == (identity.Identity{}) ||
		targetActor == "" {
		return false, acp.NewErrMissingRequiredArgToDeleteDocActorRelationship(
			policyID,
			resourceName,
			docID,
			relation,
			requestActor.DID,
			targetActor,
		)
	}

	recordFound, err := a.clientACP.DeleteActorRelationship(
		ctx,
		policyID,
		resourceName,
		docID,
		relation,
		requestActor,
		targetActor,
		protoTypes.TimestampNow(),
	)

	if err != nil {
		return false, acp.NewErrFailedToDeleteDocActorRelationshipWithACP(
			err,
			"Local",
			policyID,
			resourceName,
			docID,
			relation,
			requestActor.DID,
			targetActor,
		)
	}

	log.InfoContext(
		ctx,
		"Document and actor relationship delete",
		corelog.Any("PolicyID", policyID),
		corelog.Any("ResourceName", resourceName),
		corelog.Any("DocID", docID),
		corelog.Any("Relation", relation),
		corelog.Any("RequestActor", requestActor.DID),
		corelog.Any("TargetActor", targetActor),
		corelog.Any("RecordFound", recordFound),
	)

	return recordFound, nil
}

func (a *bridgeDocumentACP) SupportsP2P() bool {
	return a.supportsP2P
}

func (a *bridgeDocumentACP) Close() error {
	return a.clientACP.Close()
}

func (a *bridgeDocumentACP) ResetState(ctx context.Context) error {
	return a.clientACP.ResetState(ctx)
}
