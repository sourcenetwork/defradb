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
	"encoding/json"

	protoTypes "github.com/cosmos/gogoproto/types"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/acp/identity"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/acp/nac"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// NodeACPDesc contains node acp specific state information.
type NodeACPDesc struct {
	// Status represents the current state of Node ACP.
	Status client.NACStatus

	// Policy contains the policy information of the current node acp setup.
	//
	// When node access control is in a cleaned state, there will be no policy information.
	//
	// Note: The policy information must be validated at the step that enables node access the
	// very first time to ensure that the registered policy with the node acp system is valid.
	// For example, ensure that it adheres to the resource interface rules for node access control.
	Policy immutable.Option[client.PolicyDescription]
}

// NewNodeACPDesc returns a new [NodeACPDesc] that represents a clean node acp state.
func NewNodeACPDesc() NodeACPDesc {
	return NodeACPDesc{
		Status: client.NACNotConfigured,
		Policy: immutable.None[client.PolicyDescription](),
	}
}

// NACInfo contains the current node acp state information, along with the node acp instance.
type NACInfo struct {
	// NodeACP is the acp system, that is always initialized and started ([Start()] called).
	// The reason for having the system always available is to accommodate edge cases where we
	// need node access control internally even when the admin user might have had disabled it.
	// For example, when node acp was enabled once, but the admin user disabled it temporarily, then
	// to know if the identity that is re-enabling is authorized or not, we need the access control.
	NodeACP *nac.NodeACP

	// NodeACPDesc contains the current node acp specific state and other information.
	NodeACPDesc NodeACPDesc

	// EnabledInConfig is true if specified flag to start node access control for the first time.
	//
	// Note: If node access control is temporarily disabled or is already started, then this
	// config value takes no effect, and is ignored.
	EnabledInConfig bool
}

// NewNACInfo returns a newly contructed [NACInfo] with a clean [NodeACPDesc] state.
func NewNACInfo(ctx context.Context, path string, enabled bool) (NACInfo, error) {
	nodeACP, err := nac.NewNodeACP(path)
	if err != nil {
		return NACInfo{}, err
	}
	// We keep NAC started to have access control ability even when node acp is disabled
	// temporarily as we want to only allow authorized user(s) to re-enable node acp.
	if err := nodeACP.Start(ctx); err != nil {
		return NACInfo{}, err
	}

	nacInfo := NACInfo{
		NodeACP:         &nodeACP,
		NodeACPDesc:     NewNodeACPDesc(),
		EnabledInConfig: enabled,
	}
	return nacInfo, nil
}

func (db *DB) initializeNodeACP(ctx context.Context, txn datastore.Txn) error {
	isNACEnabledInStartCmd := db.nodeACP.EnabledInConfig
	wasSetupBefore, err := txn.Systemstore().Has(ctx, keys.NewNodeACPKey().Bytes())
	if err != nil {
		return err
	}

	iden := identity.FromContext(ctx)
	hasIdentity := iden.HasValue()

	// Was never setup before so start from scratch only if enabled in starting config and has identity.
	if !wasSetupBefore {
		if !isNACEnabledInStartCmd {
			log.Info("Starting without nac (not configured/setup)")
			return nil
		}

		if !hasIdentity {
			return client.ErrCanNotStartNACWithoutIdentity
		}

		// Never setup up before (or was purged), and the start command wants to enable it with identity.
		if err := db.tryRegisterNACPolicy(ctx); err != nil {
			return err
		}

		log.Info("Starting with nac, successfully configured and enabled nac")
		return nil
	}

	// Node ACP was setup before (even if it might be temporarily turned off.) We try to recover previous
	// state of node acp (overwrites [db.nodeACP.NodeACPDesc] with recovered state).
	err = db.fetchNodeACPDesc(ctx, txn)
	if err != nil {
		return err
	}

	if db.nodeACP.NodeACPDesc.Status == client.NACEnabled {
		if isNACEnabledInStartCmd {
			log.Info("Starting with nac (was already enabled, ignoring request to configure nac at start)")
			return nil
		}
		// This is when a user restarts defradb without nac explicity enabled option, when they previously
		// already configured and have nac setup, we don't want to assume they are trying to turn it off.
		// Instead we just start defradb recovering the node acp state they left before closing, and
		// notify the user how they can disable node acp if they would like to.
		log.Info("Starting with nac (can't disable nac from start cmd, use the nac disable cmd instead)")
		return nil
	}

	// Now handle the case if nac was configured before but was temporarily disabled by the authorized admin user.
	if isNACEnabledInStartCmd {
		log.Info("Starting with nac temporarily disabled (use the nac re-enable cmd to re-enable nac)")
		return nil
	} else {
		log.Info("Starting with nac already temporarily disabled, ignoring cmd for disabling nac")
		return nil
	}
}

func (db *DB) fetchNodeACPDesc(ctx context.Context, txn datastore.Txn) error {
	storedBytes, err := txn.Systemstore().Get(ctx, keys.NewNodeACPKey().Bytes())
	if err != nil {
		return err
	}

	storedNodeACPDesc := NodeACPDesc{}
	err = json.Unmarshal(storedBytes, &storedNodeACPDesc)
	if err != nil {
		return err
	}

	db.nodeACP.NodeACPDesc = storedNodeACPDesc
	return nil
}

func (db *DB) resetNodeACP(ctx context.Context) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = db.nodeACP.NodeACP.ResetState(ctx)
	if err != nil {
		return err
	}

	err = txn.Systemstore().Delete(ctx, keys.NewNodeACPKey().Bytes())
	if err != nil {
		return err
	}

	err = txn.Commit(ctx)
	if err != nil {
		return err
	}

	// Update state, only when commit is successful.
	db.nodeACP.NodeACPDesc = NewNodeACPDesc()
	return nil
}

func (db *DB) saveNodeACPDesc(ctx context.Context) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	nodeDescBytes, err := json.Marshal(db.nodeACP.NodeACPDesc)
	if err != nil {
		return err
	}

	err = txn.Systemstore().Set(ctx, keys.NewNodeACPKey().Bytes(), nodeDescBytes)
	if err != nil {
		return err
	}

	err = txn.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

// tryRegisterNACPolicy will register policy and then register the node with the node acp system,
// if anything goes wrong, it will leave the node acp in the clean state.
// For example if uploading policy succeeds but registering fails, then purge/reset the state to
// not have that policy floating there (as it can cause a corrupt state).
//
// Upon success returns nil, and modifies [db.nodeACP] appropriately.
// Upon failure returns an error and ensures clean node acp state.
//
// Note:
// - This function should only be called when starting node acp from a clean state.
func (db *DB) tryRegisterNACPolicy(ctx context.Context) error {
	iden := identity.FromContext(ctx)
	if !iden.HasValue() {
		return ErrNoIdentityInContext
	}
	identity := iden.Value()

	// Having a non-empty identity is a MUST requirement for adding a policy.
	if identity == nil || identity.DID() == "" {
		return acp.ErrPolicyCreatorMustNotBeEmpty
	}

	// Must have have node acp instance setup.
	if db.nodeACP.NodeACP == nil {
		return ErrNACIsEnabledButInstanceIsNotAvailable
	}

	policyID, err := db.nodeACP.NodeACP.AddPolicy(
		ctx,
		identity,
		acpTypes.NodeACPPolicy,
		acpTypes.PolicyMarshalType_YAML,
		protoTypes.TimestampNow(),
	)
	if err != nil {
		return err
	}

	// Validate the policy is valid according to node acp resource interface rules.
	// Issue: https://github.com/sourcenetwork/defradb/issues/3718
	// TODO: Maybe move this check before uploading policy, so upon failure we don't have
	// to reset. If above is not possible in a clean way, we can atleast consolodate nac
	// and dac `ValidateResourceInterface` implementations. We don't have to worry about
	// this too much right now as we are guaranteed the internal policy will always be valid.
	err = db.nodeACP.NodeACP.ValidateResourceInterface(
		ctx,
		policyID,
		acpTypes.NodeACPPolicyResourceName,
	)
	if err != nil { // We must fix the state before returning, as we already uploaded the policy.
		if errReset := db.resetNodeACP(ctx); errReset != nil {
			return errors.Join(errReset, err)
		}
		return err
	}

	err = db.nodeACP.NodeACP.RegisterObject(
		ctx,
		identity,
		policyID,
		acpTypes.NodeACPPolicyResourceName,
		acpTypes.NodeACPObject,
		protoTypes.TimestampNow(),
	)
	if err != nil { // We must fix the state before returning, as we already uploaded the policy.
		if errReset := db.resetNodeACP(ctx); errReset != nil {
			return errors.Join(errReset, err)
		}
		return err
	}

	policyDesc := client.PolicyDescription{
		ID:           policyID,
		ResourceName: acpTypes.NodeACPPolicyResourceName,
	}

	db.nodeACP.NodeACPDesc.Status = client.NACEnabled
	db.nodeACP.NodeACPDesc.Policy = immutable.Some(policyDesc)
	return db.saveNodeACPDesc(ctx)
}

// ReEnableNAC will re-enable node acp that was temporarily disabled (and configured). This will
// recover previously saved nac state with all the relationships formed.
//
// If node acp is already enabled, then returns an error reflecting that it is already enabled.
//
// If node acp is not already configured or the previous state was purged then this will return an error,
// as the user must use the node's start command to configure/enable the node acp the first time.
//
// Returns an [client.ErrNotAuthorizedToPerformOperation] error if the requesting identity is not
// authorized to perform this operation.
func (db *DB) ReEnableNAC(ctx context.Context) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if db.nodeACP.NodeACPDesc.Status == client.NACNotConfigured {
		return ErrNACIsNotConfigured
	}

	if db.nodeACP.NodeACPDesc.Status == client.NACEnabled {
		return ErrNACIsAlreadyEnabled
	}

	// User trying to re-enable a disabled nac state.
	// Check if this request is authorized to re-enable node access control.
	if err := db.checkNodeAccess(ctx, acpTypes.NodeNACReEnablePerm); err != nil {
		return err
	}

	db.nodeACP.NodeACPDesc.Status = client.NACEnabled
	return db.saveNodeACPDesc(ctx)
}

// DisableNAC will disable node acp for the users temporarily. This will keep the current node acp
// state saved so that if it is re-enabled in the future, then we can recover all the relationships formed.
//
// If node acp is already disabled, then returns an error reflecting that it is already disabled.
//
// If node acp is not already configured or the previous state was purged then this will return an error.
//
// Returns an [client.ErrNotAuthorizedToPerformOperation] error if the requesting identity is not
// authorized to perform this operation.
func (db *DB) DisableNAC(ctx context.Context) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if db.nodeACP.NodeACPDesc.Status == client.NACNotConfigured {
		return ErrNACIsNotConfigured
	}

	if db.nodeACP.NodeACPDesc.Status == client.NACDisabledTemporarily {
		return ErrNACIsAlreadyDisabled
	}

	// Check if this request is authorized to disable node access control.
	if err := db.checkNodeAccess(ctx, acpTypes.NodeNACDisablePerm); err != nil {
		return err
	}

	db.nodeACP.NodeACPDesc.Status = client.NACDisabledTemporarily
	return db.saveNodeACPDesc(ctx)
}

// PurgeNACState will reset/purge the entire node acp state. This means that all relationships that were formed
// will be deleted and any user can then enable node acp using their identity and become the admin user (owner).
//
// Returns an [client.ErrNotAuthorizedToPerformOperation] error if the requesting identity is not
// authorized to perform this operation.
//
// Note:
// - This will disable node acp and leave it in a clean state.
// - This operation also requires dev-mode to be enabled.
func (db *DB) PurgeNACState(ctx context.Context) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if db.nodeACP.NodeACP != nil {
		err := db.resetNodeACP(ctx)
		if err != nil {
			log.ErrorE("Failed to reset node ACP state", err)
			return err
		}
	}

	return nil
}

// checkNodeAccess is a helper function that performs the node acp validation check, if requesting
// user has access then nil is returned, otherwise returns an error.
//
// Returns an [client.ErrNotAuthorizedToPerformOperation] error if the requesting identity is not
// authorized to perform this operation.
//
// Note:
// - If the requesting identity is the nodeIdentity then we assume it has access.
// - If the operation needs the nac permission to execute, it must have nac configured (not clean).
func (db *DB) checkNodeAccess(
	ctx context.Context,
	permissionNeeded acpTypes.NodeResourcePermission,
) error {
	// For nac specific operations, the node acp setup must be configured.
	if permissionNeeded.IsForNACOperation() &&
		db.nodeACP.NodeACPDesc.Status == client.NACNotConfigured &&
		permissionNeeded != acpTypes.NodeNACStatusPerm {
		return ErrNACIsNotConfigured
	}

	ident := identity.FromContext(ctx)
	if ident.HasValue() &&
		db.nodeIdentity.HasValue() &&
		ident.Value().DID() == db.nodeIdentity.Value().DID() {
		return nil
	}

	return CheckNodeOperationAccess(
		ctx,
		ident,
		db.nodeACP,
		permissionNeeded,
		acpTypes.NodeACPObject,
	)
}

// CheckNodeOperationAccess returns an [client.ErrNotAuthorizedToPerformOperation]
// error if the requesting user does not have the required permission to perform an operation.
// If something else goes wrong, it returns a different error, otherwise returns nil only if
// the check passes and the requesting user is authorized to perform the operation.
//
// Unrestricted access if:
// - node acp system is temporarily disabled (unless the operation is trying to turn on nac).
func CheckNodeOperationAccess(
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	nacInfo NACInfo,
	permission acpTypes.ResourceInterfacePermission,
	objectID string,
) error {
	if nacInfo.NodeACPDesc.Status != client.NACEnabled &&
		permission != acpTypes.NodeNACReEnablePerm {
		// Unrestricted access if node acp is off, and not trying to turn it back on.
		return nil
	}

	// If node acp is enabled then it must have have node acp instance setup.
	if nacInfo.NodeACP == nil {
		return ErrNACIsEnabledButInstanceIsNotAvailable
	}

	// If node acp is enabled then it must have a valid policy information.
	if !nacInfo.NodeACPDesc.Policy.HasValue() {
		return ErrNACIsEnabledButIsMissingPolicyInfo
	}

	policyID := nacInfo.NodeACPDesc.Policy.Value().ID
	resourceName := nacInfo.NodeACPDesc.Policy.Value().ResourceName
	if policyID == "" || resourceName == "" {
		return ErrNACIsEnabledButIsMissingPolicyInfo
	}
	// Since public node will have unrestricted access, the object we are gating MUST be registered
	// if node access control is configured.
	isRegistered, err := nacInfo.NodeACP.ObjectOwner(
		ctx,
		policyID,
		resourceName,
		objectID,
	)
	if err != nil {
		return err
	}

	if !isRegistered.HasValue() {
		return ErrNACNodeObjectToGateIsNotRegistered
	}

	var identityValue string
	if !identity.HasValue() {
		// We can't assume that there is no-access just because there is no identity even if the operation
		// is registered with acp, this is because it is possible that acp has a registered relation targeting
		// "*" (any) actor which would mean that even a request without an identity might be able to access
		// an operation registered with acp. So we pass an empty `did` to accommodate that case.
		identityValue = ""
	} else {
		identityValue = identity.Value().DID()
	}

	nodeResourcePerm, ok := permission.(acpTypes.NodeResourcePermission)
	if !ok {
		return client.ErrInvalidResourcePermissionType
	}

	// Now actually check if this identity has access or not.
	hasAccess, err := nacInfo.NodeACP.VerifyAccessRequest(
		ctx,
		nodeResourcePerm,
		identityValue,
		policyID,
		resourceName,
		objectID,
	)

	if err != nil {
		return acp.NewErrFailedToVerifyNodeAccessWithACP(
			err,
			permission.String(),
			policyID,
			identityValue,
			resourceName,
			objectID,
		)
	}

	if hasAccess {
		return nil
	}

	return client.ErrNotAuthorizedToPerformOperation
}

func (db *DB) GetNACStatus(ctx context.Context) (client.NACStatusResult, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if err := db.checkNodeAccess(ctx, acpTypes.NodeNACStatusPerm); err != nil {
		return client.NACStatusResult{}, err
	}

	return client.NACStatusResult{
		Status: db.nodeACP.NodeACPDesc.Status.String(),
	}, nil
}

func (db *DB) AddNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if err := db.checkNodeAccess(ctx, acpTypes.NodeNACRelationAddPerm); err != nil {
		return client.AddActorRelationshipResult{}, err
	}

	return db.addNACActorRelationship(ctx, relation, targetActor)
}

func (db *DB) addNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	// Note: While we give unrestricted access when node acp if turned off, there are certain
	// requests that we can't do when node acp is turned off or unavailable, this is one of them.
	if db.nodeACP.NodeACPDesc.Status != client.NACEnabled ||
		db.nodeACP.NodeACP == nil {
		return client.AddActorRelationshipResult{}, client.ErrACPOperationButACPNotAvailable
	}

	if !db.nodeACP.NodeACPDesc.Policy.HasValue() {
		return client.AddActorRelationshipResult{}, ErrNACIsEnabledButIsMissingPolicyInfo
	}

	policyDesc := db.nodeACP.NodeACPDesc.Policy.Value()
	if policyDesc.ID == "" || policyDesc.ResourceName == "" {
		return client.AddActorRelationshipResult{}, ErrNACIsEnabledButIsMissingPolicyInfo
	}

	requestActor := identity.FromContext(ctx)
	if !requestActor.HasValue() || requestActor.Value() == nil || requestActor.Value().DID() == "" {
		return client.AddActorRelationshipResult{}, ErrNACRelationshipOperationRequiresIdentity
	}

	exists, err := db.nodeACP.NodeACP.AddActorRelationship(
		ctx,
		db.nodeACP.NodeACPDesc.Policy.Value().ID,
		db.nodeACP.NodeACPDesc.Policy.Value().ResourceName,
		acpTypes.NodeACPObject,
		relation,
		requestActor.Value(),
		targetActor,
		protoTypes.TimestampNow(),
	)
	if err != nil {
		return client.AddActorRelationshipResult{}, err
	}

	return client.AddActorRelationshipResult{ExistedAlready: exists}, nil
}

func (db *DB) DeleteNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if err := db.checkNodeAccess(ctx, acpTypes.NodeNACRelationDeletePerm); err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}

	return db.deleteNACActorRelationship(ctx, relation, targetActor)
}

func (db *DB) deleteNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	// Note: While we give unrestricted access when node acp if turned off, there are certain
	// requests that we can't do when node acp is turned off or unavailable, this is one of them.
	if db.nodeACP.NodeACPDesc.Status != client.NACEnabled ||
		db.nodeACP.NodeACP == nil {
		return client.DeleteActorRelationshipResult{}, client.ErrACPOperationButACPNotAvailable
	}

	if !db.nodeACP.NodeACPDesc.Policy.HasValue() {
		return client.DeleteActorRelationshipResult{}, ErrNACIsEnabledButIsMissingPolicyInfo
	}

	policyDesc := db.nodeACP.NodeACPDesc.Policy.Value()
	if policyDesc.ID == "" || policyDesc.ResourceName == "" {
		return client.DeleteActorRelationshipResult{}, ErrNACIsEnabledButIsMissingPolicyInfo
	}

	requestActor := identity.FromContext(ctx)
	if !requestActor.HasValue() || requestActor.Value() == nil || requestActor.Value().DID() == "" {
		return client.DeleteActorRelationshipResult{}, ErrNACRelationshipOperationRequiresIdentity
	}

	recordFound, err := db.nodeACP.NodeACP.DeleteActorRelationship(
		ctx,
		policyDesc.ID,
		policyDesc.ResourceName,
		acpTypes.NodeACPObject,
		relation,
		requestActor.Value(),
		targetActor,
		protoTypes.TimestampNow(),
	)
	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}

	return client.DeleteActorRelationshipResult{RecordFound: recordFound}, nil
}
