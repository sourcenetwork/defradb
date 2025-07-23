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
	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/acp/aac"
	"github.com/sourcenetwork/defradb/acp/identity"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// AdminDesc contains admin acp specific state information.
type AdminACPDesc struct {
	// Status represents the current state of Node ACP.
	Status client.NACStatus

	// Policy contains the policy information of the current admin acp setup.
	//
	// When admin access control is in a cleaned state, there will be no policy information.
	//
	// Note: The policy information must be validated at the step that enables admin access the
	// very first time to ensure that the registered policy with the admin acp system is valid.
	// For example, ensure that it adheres to the resource interface rules for admin access control.
	Policy immutable.Option[client.PolicyDescription]
}

// NewAdminACPDesc returns a new [AdminACPDesc] that represents a clean admin acp state.
func NewAdminACPDesc() AdminACPDesc {
	return AdminACPDesc{
		Status: client.NACNotConfigured,
		Policy: immutable.None[client.PolicyDescription](),
	}
}

// AdminInfo contains the current admin acp state information, along with the admin acp instance.
type AdminInfo struct {
	// AdminACP is the acp system, that is always initialized and started ([Start()] called).
	// The reason for having the system always available is to accommodate edge cases where we
	// need admin access control internally even when the admin might have had disabled it.
	// For example, when admin acp was enabled once, but the admin disabled it temporarily, then
	// to know if the identity that is re-enabling is authorized or not, we need the access control.
	//
	// Note:
	// - Check [AdminDesc.IsEnabled] to know if admin access control is enabled or disabled.
	// - Check [AdminDesc.IsClean] to know if admin access control was ever enabled before.
	AdminACP *aac.AdminACP

	// AdminDesc contains the current admin acp specific state and other information.
	AdminDesc AdminACPDesc
}

// NewCleanAdminInfo returns a newly contructed [AdminInfo] with a clean [AdminDesc] state.
func NewCleanAdminInfo() AdminInfo {
	adminACP := aac.NewAdminACP()
	adminInfo := AdminInfo{
		AdminACP:  &adminACP,
		AdminDesc: NewCleanAdminACPDesc(),
	}
	return adminInfo
}

// NewAdminInfoWithAACDisabled returns an [AdminInfo] with admin acp system disabled in path.
// Returns an error if initialization failed.
//
// Note: Caller is responsible for calling [AdminInfo.AdminACP.Close()] to free resources.
func NewAdminInfoWithAACDisabled(ctx context.Context, path string) (AdminInfo, error) {
	adminInfo := NewCleanAdminInfo()
	adminInfo.AdminACP.Init(ctx, path)

	// We keep AAC started to have access control ability even when admin acp is disabled
	// temporarily as we want to only allow authorized user(s) to re-enable admin acp.
	if err := adminInfo.AdminACP.Start(ctx); err != nil {
		return AdminInfo{}, err
	}

	return adminInfo, nil
}

// NewAdminInfoWithAACEnabled returns an [AdminInfo] with admin acp system enabled in path.
// Returns an error if initialization failed.
//
// Note: Caller is responsible for calling [AdminInfo.AdminACP.Close()] to free resources.
func NewAdminInfoWithAACEnabled(ctx context.Context, path string) (AdminInfo, error) {
	adminInfo, err := NewAdminInfoWithAACDisabled(ctx, path)
	if err != nil {
		return AdminInfo{}, err
	}
	adminInfo.AdminDesc.IsEnabled = true
	return adminInfo, nil
}

func (db *DB) initializeAdminACP(ctx context.Context, txn datastore.Txn) error {
	isAACEnabledInStartCmd := db.adminInfo.AdminDesc.IsEnabled
	wasSetupBefore, err := txn.Systemstore().Has(ctx, keys.NewAdminACPKey().Bytes())
	if err != nil && !errors.Is(err, corekv.ErrNotFound) {
		return err
	}

	iden := identity.FromContext(ctx)
	hasIdentity := iden.HasValue()

	// Was never setup before so start from scratch only if enabled in starting config and has identity.
	if !wasSetupBefore {
		if !isAACEnabledInStartCmd {
			log.Info("Starting without aac (not configured/setup)")
			return nil
		}

		if !hasIdentity {
			return client.ErrCanNotStartAACWithoutIdentity
		}

		// Never setup up before (or was purged), and the start command wants to enable it with identity.
		if err := db.cleanRegisterInternalAACPolicyAndGateNode(ctx); err != nil {
			return err
		}

		log.Info("Starting with aac, successfully configured and enabled aac")
		return nil
	}

	// Admin ACP was setup before (even if it might be temporarily turned off.) We try to recover previous
	// state of admin acp (overwrites [db.adminInfo.AdminDesc] with recovered state).
	err = db.fetchAdminACPDesc(ctx, txn)
	if err != nil {
		return err
	}

	if db.adminInfo.AdminDesc.Status == client.NACEnabled {
		if isAACEnabledInStartCmd {
			log.Info("Starting with aac (was already enabled, ignoring request to configure aac at start)")
			return nil
		}
		// This is when a user restarts defradb without aac explicity enabled option, when they previously
		// already configured and have aac setup, we don't want to assume they are trying to turn it off.
		// Instead we just start defradb recovering the admin acp state they left before closing, and
		// notify the user how they can disable admin acp if they would like to.
		log.Info("Starting with aac (can't disable aac from start cmd, use the aac disable cmd instead)")
		return nil
	}

	// Now handle the case if aac was configured before but was temporarily disabled by the authorized admin.
	if isAACEnabledInStartCmd {
		log.Info("Starting with aac temporarily disabled (use the aac re-enable cmd to re-enable aac)")
		return nil
	} else {
		log.Info("Starting with aac already temporarily disabled, ignoring cmd for disabling aac")
		return nil
	}
}

func (db *DB) fetchAdminACPDesc(ctx context.Context, txn datastore.Txn) error {
	storedBytes, err := txn.Systemstore().Get(ctx, keys.NewAdminACPKey().Bytes())
	if err != nil {
		return err
	}

	storedAdminACPDesc := AdminACPDesc{}
	err = json.Unmarshal(storedBytes, &storedAdminACPDesc)
	if err != nil {
		return err
	}

	db.adminInfo.AdminDesc = storedAdminACPDesc
	return nil
}

func (db *DB) resetAdminACP(ctx context.Context) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = db.adminInfo.AdminACP.ResetState(ctx)
	if err != nil {
		return err
	}

	err = txn.Systemstore().Delete(ctx, keys.NewAdminACPKey().Bytes())
	if err != nil {
		return err
	}

	err = txn.Commit(ctx)
	if err != nil {
		return err
	}

	// Update state, only when commit is successful.
	db.adminInfo.AdminDesc = NewAdminACPDesc()
	return nil
}

func (db *DB) saveAdminACPDesc(ctx context.Context) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	adminDescBytes, err := json.Marshal(db.adminInfo.AdminDesc)
	if err != nil {
		return err
	}

	err = txn.Systemstore().Set(ctx, keys.NewAdminACPKey().Bytes(), adminDescBytes)
	if err != nil {
		return err
	}

	err = txn.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

// cleanRegisterInternalAACPolicyAndGateNode will register policy and then register the node with the
// admin acp system, if anything goes wrong, it will leave the admin acp in the clean state.
// For example if uploading policy succeeds but registering fails, then purge/reset the state to
// not have that policy floating there (as it can cause a corrupt state).
//
// Upon success returns nil, and modifies [db.adminInfo] appropriately.
// Upon failure returns an error and ensures clean admin acp state.
//
// Note:
// - This function should only be called when starting admin acp from a clean state.
func (db *DB) cleanRegisterInternalAACPolicyAndGateNode(ctx context.Context) error {
	iden := identity.FromContext(ctx)
	if !iden.HasValue() {
		return ErrNoIdentityInContext
	}
	identity := iden.Value()

	// Having a non-empty identity is a MUST requirement for adding a policy.
	if identity == nil || identity.DID() == "" {
		return acp.ErrPolicyCreatorMustNotBeEmpty
	}

	// Must have have admin acp instance setup.
	if db.adminInfo.AdminACP == nil {
		return ErrAACIsEnabledButInstanceIsNotAvailable
	}

	policyID, err := db.adminInfo.AdminACP.AddPolicy(
		ctx,
		identity,
		acpTypes.NodeACPPolicy,
		acpTypes.PolicyMarshalType_YAML,
		protoTypes.TimestampNow(),
	)
	if err != nil {
		return err
	}

	// Validate the policy is valid according to admin acp resource interface rules.
	// Issue: https://github.com/sourcenetwork/defradb/issues/3718
	// TODO: Maybe move this check before uploading policy, so upon failure we don't have
	// to reset. If above is not possible in a clean way, we can atleast consolodate aac
	// and dac `ValidateResourceInterface` implementations. We don't have to worry about
	// this too much right now as we are guaranteed the internal policy will always be valid.
	err = db.adminInfo.AdminACP.ValidateResourceInterface(
		ctx,
		policyID,
		acpTypes.NodeACPPolicyResourceName,
	)
	if err != nil { // We must fix the state before returning, as we already uploaded the policy.
		if errReset := db.resetAdminACP(ctx); errReset != nil {
			return errors.Join(errReset, err)
		}
		return err
	}

	err = db.adminInfo.AdminACP.RegisterObject(
		ctx,
		identity,
		policyID,
		acpTypes.NodeACPPolicyResourceName,
		acpTypes.NodeACPObject,
		protoTypes.TimestampNow(),
	)
	if err != nil { // We must fix the state before returning, as we already uploaded the policy.
		if errReset := db.resetAdminACP(ctx); errReset != nil {
			return errors.Join(errReset, err)
		}
		return err
	}

	policyDesc := client.PolicyDescription{
		ID:           policyID,
		ResourceName: acpTypes.NodeACPPolicyResourceName,
	}

	db.adminInfo.AdminDesc.Status = client.NACEnabled
	db.adminInfo.AdminDesc.Policy = immutable.Some(policyDesc)
	return db.saveAdminACPDesc(ctx)
}

// ReEnableAAC will re-enable admin acp that was temporarily disabled (and configured). This will
// recover previously saved aac state with all the relationships formed.
//
// If admin acp is already enabled, then returns an error reflecting that it is already enabled.
//
// If admin acp is not already configured or the previous state was purged then this will return an error,
// as the user must use the node's start command to configure/enable the admin acp the first time.
//
// Returns an [client.ErrNotAuthorizedToPerformOperation] error if the requesting identity is not
// authorized to perform this operation.
func (db *DB) ReEnableAAC(ctx context.Context) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if db.adminInfo.AdminDesc.Status == client.NACNotConfigured {
		return ErrAACIsNotConfigured
	}

	if db.adminInfo.AdminDesc.Status == client.NACEnabled {
		return ErrAACIsAlreadyEnabled
	}

	// User trying to re-enable a disabled aac state.
	// Check if this request is authorized to re-enable admin access control.
	if err := db.checkAdminAccess(ctx, acpTypes.AdminAACReEnablePerm); err != nil {
		return err
	}

	db.adminInfo.AdminDesc.Status = client.NACEnabled
	return db.saveAdminACPDesc(ctx)
}

// DisableAAC will disable admin acp for the users temporarily. This will keep the current admin acp
// state saved so that if it is re-enabled in the future, then we can recover all the relationships formed.
//
// If admin acp is already disabled, then returns an error reflecting that it is already disabled.
//
// If admin acp is not already configured or the previous state was purged then this will return an error.
//
// Returns an [client.ErrNotAuthorizedToPerformOperation] error if the requesting identity is not
// authorized to perform this operation.
func (db *DB) DisableAAC(ctx context.Context) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if db.adminInfo.AdminDesc.Status == client.NACNotConfigured {
		return ErrAACIsNotConfigured
	}

	if db.adminInfo.AdminDesc.Status == client.NACDisabledTemporarily {
		return ErrAACIsAlreadyDisabled
	}

	// Check if this request is authorized to disable admin access control.
	if err := db.checkAdminAccess(ctx, acpTypes.AdminAACDisablePerm); err != nil {
		return err
	}

	db.adminInfo.AdminDesc.Status = client.NACDisabledTemporarily
	return db.saveAdminACPDesc(ctx)
}

// checkAdminAccess is a helper function that performs the admin acp validation check, if requesting
// user has access then nil is returned, otherwise returns an error.
//
// Returns an [client.ErrNotAuthorizedToPerformOperation] error if the requesting identity is not
// authorized to perform this operation.
//
// Note:
// - If the requesting identity is the nodeIdentity then we assume it has access.
// - If the operation needs the aac permission to execute, it must have aac configured (not clean).
func (db *DB) checkAdminAccess(
	ctx context.Context,
	permissionNeeded acpTypes.AdminResourcePermission,
) error {
	// For aac specific operations, the admin acp setup must be configured.
	if permissionNeeded.IsForAACOperation() &&
		db.adminInfo.AdminDesc.Status == client.NACNotConfigured &&
		permissionNeeded != acpTypes.AdminAACStatusPerm {
		return ErrAACIsNotConfigured
	}

	ident := identity.FromContext(ctx)
	if ident.HasValue() &&
		db.nodeIdentity.HasValue() &&
		ident.Value().DID() == db.nodeIdentity.Value().DID() {
		return nil
	}

	return CheckAACNodeOperationAccess(
		ctx,
		ident,
		db.adminInfo,
		permissionNeeded,
		acpTypes.NodeACPObject,
	)
}

// CheckAACNodeOperationAccess returns an [client.ErrNotAuthorizedToPerformOperation]
// error if the requesting user does not have the required permission to perform an operation.
// If something else goes wrong, it returns a different error, otherwise returns nil only if
// the check passes and the requesting user is authorized to perform the operation.
//
// Unrestricted access if:
// - admin acp system is temporarily disabled (unless the operation is trying to turn it on aac).
func CheckAACNodeOperationAccess(
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	adminInfo AdminInfo,
	permission acpTypes.ResourceInterfacePermission,
	objectID string,
) error {
	if adminInfo.AdminDesc.Status != client.NACEnabled &&
		permission != acpTypes.AdminAACReEnablePerm {
		// Unrestricted access if admin acp is off, and not trying to turn it back on.
		return nil
	}

	// If admin acp is enabled then it must have have admin acp instance setup.
	if adminInfo.AdminACP == nil {
		return ErrAACIsEnabledButInstanceIsNotAvailable
	}

	// If admin acp is enabled then it must have a valid policy information.
	if !adminInfo.AdminDesc.Policy.HasValue() {
		return ErrAACIsEnabledButIsMissingPolicyInfo
	}

	policyID := adminInfo.AdminDesc.Policy.Value().ID
	resourceName := adminInfo.AdminDesc.Policy.Value().ResourceName
	if policyID == "" || resourceName == "" {
		return ErrAACIsEnabledButIsMissingPolicyInfo
	}
	// Since public node will have unrestricted access, the object we are gating MUST be registered
	// if admin access control is configured.
	isRegistered, err := adminInfo.AdminACP.ObjectOwner(
		ctx,
		policyID,
		resourceName,
		objectID,
	)
	if err != nil {
		return err
	}

	if !isRegistered.HasValue() {
		return ErrAACNodeObjectToGateIsNotRegistered
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

	adminResourcePerm, ok := permission.(acpTypes.AdminResourcePermission)
	if !ok {
		return client.ErrInvalidResourcePermissionType
	}

	// Now actually check if this identity has access or not.
	hasAccess, err := adminInfo.AdminACP.VerifyAccessRequest(
		ctx,
		adminResourcePerm,
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

func (db *DB) GetAACStatus(ctx context.Context) (client.StatusAACResult, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if err := db.checkAdminAccess(ctx, acpTypes.AdminAACStatusPerm); err != nil {
		return client.StatusAACResult{}, err
	}

	return client.StatusAACResult{
		Status: db.adminInfo.AdminDesc.Status.String(),
	}, nil
}

func (db *DB) AddAACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if err := db.checkAdminAccess(ctx, acpTypes.AdminAACRelationAddPerm); err != nil {
		return client.AddActorRelationshipResult{}, err
	}

	return db.addAACActorRelationship(ctx, relation, targetActor)
}

func (db *DB) addAACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	// Note: While we give unrestricted access when admin acp if turned off, there are certain
	// requests that we can't do when admin acp is turned off or unavailable, this is one of them.
	if db.adminInfo.AdminDesc.Status != client.NACEnabled ||
		db.adminInfo.AdminACP == nil {
		return client.AddActorRelationshipResult{}, client.ErrACPOperationButACPNotAvailable
	}

	if !db.adminInfo.AdminDesc.Policy.HasValue() {
		return client.AddActorRelationshipResult{}, ErrAACIsEnabledButIsMissingPolicyInfo
	}

	policyDesc := db.adminInfo.AdminDesc.Policy.Value()
	if policyDesc.ID == "" || policyDesc.ResourceName == "" {
		return client.AddActorRelationshipResult{}, ErrAACIsEnabledButIsMissingPolicyInfo
	}

	requestActor := identity.FromContext(ctx)
	if !requestActor.HasValue() || requestActor.Value() == nil || requestActor.Value().DID() == "" {
		return client.AddActorRelationshipResult{}, ErrAACRelationshipOperationRequiresIdentity
	}

	exists, err := db.adminInfo.AdminACP.AddActorRelationship(
		ctx,
		db.adminInfo.AdminDesc.Policy.Value().ID,
		db.adminInfo.AdminDesc.Policy.Value().ResourceName,
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

func (db *DB) DeleteAACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if err := db.checkAdminAccess(ctx, acpTypes.AdminAACRelationDeletePerm); err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}

	return db.deleteAACActorRelationship(ctx, relation, targetActor)
}

func (db *DB) deleteAACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	// Note: While we give unrestricted access when admin acp if turned off, there are certain
	// requests that we can't do when admin acp is turned off or unavailable, this is one of them.
	if db.adminInfo.AdminDesc.Status != client.NACEnabled ||
		db.adminInfo.AdminACP == nil {
		return client.DeleteActorRelationshipResult{}, client.ErrACPOperationButACPNotAvailable
	}

	if !db.adminInfo.AdminDesc.Policy.HasValue() {
		return client.DeleteActorRelationshipResult{}, ErrAACIsEnabledButIsMissingPolicyInfo
	}

	policyDesc := db.adminInfo.AdminDesc.Policy.Value()
	if policyDesc.ID == "" || policyDesc.ResourceName == "" {
		return client.DeleteActorRelationshipResult{}, ErrAACIsEnabledButIsMissingPolicyInfo
	}

	requestActor := identity.FromContext(ctx)
	if !requestActor.HasValue() || requestActor.Value() == nil || requestActor.Value().DID() == "" {
		return client.DeleteActorRelationshipResult{}, ErrAACRelationshipOperationRequiresIdentity
	}

	recordFound, err := db.adminInfo.AdminACP.DeleteActorRelationship(
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
