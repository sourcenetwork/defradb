// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package acp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/acp/identity"
)

var identity1 = identity.Identity{
	DID: "did:key:z7r8os2G88XXBNBTLj3kFR5rzUJ4VAesbX7PgsA68ak9B5RYcXF5EZEmjRzzinZndPSSwujXb4XKHG6vmKEFG6ZfsfcQn",
}
var identity2 = identity.Identity{
	DID: "did:key:z7r8ooUiNXK8TT8Xjg1EWStR2ZdfxbzVfvGWbA2FjmzcnmDxz71QkP1Er8PP3zyLZpBLVgaXbZPGJPS4ppXJDPRcqrx4F",
}
var invalidIdentity = identity.Identity{
	DID: "did:something",
}

var validPolicyID string = "d59f91ba65fe142d35fc7df34482eafc7e99fed7c144961ba32c4664634e61b7"
var validPolicy string = `
name: test
description: a policy

actor:
  name: actor

resources:
  users:
    permissions:
      write:
        expr: owner
      read:
        expr: owner + reader

    relations:
      owner:
        types:
          - actor
      reader:
        types:
          - actor
 `

func Test_LocalACP_InMemory_StartAndClose_NoError(t *testing.T) {
	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, "")
	err := localACP.Start(ctx)

	require.Nil(t, err)

	err = localACP.Close()
	require.Nil(t, err)
}

func Test_LocalACP_PersistentMemory_StartAndClose_NoError(t *testing.T) {
	acpPath := t.TempDir()
	require.NotEqual(t, "", acpPath)

	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, acpPath)
	err := localACP.Start(ctx)
	require.Nil(t, err)

	err = localACP.Close()
	require.Nil(t, err)
}

func Test_LocalACP_InMemory_AddPolicy_CreatingSamePolicyAfterWipeReturnsSameID(t *testing.T) {
	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, "")
	errStart := localACP.Start(ctx)
	require.Nil(t, errStart)

	policyID, errAddPolicy := localACP.AddPolicy(
		ctx,
		identity1,
		validPolicy,
	)
	require.Nil(t, errAddPolicy)

	require.Equal(
		t,
		validPolicyID,
		policyID,
	)

	errClose := localACP.Close()
	require.Nil(t, errClose)

	// Since nothing is persisted should allow adding same policy again with same ID

	localACP.Init(ctx, "")
	errStart = localACP.Start(ctx)
	require.Nil(t, errStart)

	policyID, errAddPolicy = localACP.AddPolicy(
		ctx,
		identity1,
		validPolicy,
	)
	require.Nil(t, errAddPolicy)
	require.Equal(
		t,
		validPolicyID,
		policyID,
	)

	errClose = localACP.Close()
	require.Nil(t, errClose)
}

func Test_LocalACP_PersistentMemory_AddPolicy_CreatingSamePolicyReturnsDifferentIDs(t *testing.T) {
	acpPath := t.TempDir()
	require.NotEqual(t, "", acpPath)

	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, acpPath)
	errStart := localACP.Start(ctx)
	require.Nil(t, errStart)

	policyID, errAddPolicy := localACP.AddPolicy(
		ctx,
		identity1,
		validPolicy,
	)
	require.Nil(t, errAddPolicy)
	require.Equal(
		t,
		validPolicyID,
		policyID,
	)

	errClose := localACP.Close()
	require.Nil(t, errClose)

	// The above policy should remain persisted on restarting ACP.

	localACP.Init(ctx, acpPath)
	errStart = localACP.Start(ctx)
	require.Nil(t, errStart)

	// Should generate a different ID for the new policy, even though the payload is the same
	newPolicyID, errAddPolicy := localACP.AddPolicy(
		ctx,
		identity1,
		validPolicy,
	)
	require.NoError(t, errAddPolicy)
	require.NotEqual(t, newPolicyID, policyID)

	errClose = localACP.Close()
	require.Nil(t, errClose)
}

func Test_LocalACP_InMemory_ValidateResourseExistsOrNot_ErrIfDoesntExist(t *testing.T) {
	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, "")
	errStart := localACP.Start(ctx)
	require.Nil(t, errStart)

	policyID, errAddPolicy := localACP.AddPolicy(
		ctx,
		identity1,
		validPolicy,
	)
	require.Nil(t, errAddPolicy)
	require.Equal(
		t,
		validPolicyID,
		policyID,
	)

	errValidateResourceExists := localACP.ValidateResourceExistsOnValidDPI(
		ctx,
		validPolicyID,
		"users",
	)
	require.Nil(t, errValidateResourceExists)

	errValidateResourceExists = localACP.ValidateResourceExistsOnValidDPI(
		ctx,
		validPolicyID,
		"resourceDoesNotExist",
	)
	require.Error(t, errValidateResourceExists)
	require.ErrorIs(t, errValidateResourceExists, ErrResourceDoesNotExistOnTargetPolicy)

	errValidateResourceExists = localACP.ValidateResourceExistsOnValidDPI(
		ctx,
		"invalidPolicyID",
		"resourceDoesNotExist",
	)
	require.Error(t, errValidateResourceExists)
	require.ErrorIs(t, errValidateResourceExists, ErrPolicyDoesNotExistWithACP)

	errClose := localACP.Close()
	require.Nil(t, errClose)
}

func Test_LocalACP_PersistentMemory_ValidateResourseExistsOrNot_ErrIfDoesntExist(t *testing.T) {
	acpPath := t.TempDir()
	require.NotEqual(t, "", acpPath)

	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, acpPath)
	errStart := localACP.Start(ctx)
	require.Nil(t, errStart)

	policyID, errAddPolicy := localACP.AddPolicy(
		ctx,
		identity1,
		validPolicy,
	)
	require.Nil(t, errAddPolicy)
	require.Equal(
		t,
		validPolicyID,
		policyID,
	)

	errValidateResourceExists := localACP.ValidateResourceExistsOnValidDPI(
		ctx,
		validPolicyID,
		"users",
	)
	require.Nil(t, errValidateResourceExists)

	// Resource should still exist even after a restart.
	errClose := localACP.Close()
	require.Nil(t, errClose)

	localACP.Init(ctx, acpPath)
	errStart = localACP.Start(ctx)
	require.Nil(t, errStart)

	// Do the same check after restart.
	errValidateResourceExists = localACP.ValidateResourceExistsOnValidDPI(
		ctx,
		validPolicyID,
		"users",
	)
	require.Nil(t, errValidateResourceExists)

	errValidateResourceExists = localACP.ValidateResourceExistsOnValidDPI(
		ctx,
		validPolicyID,
		"resourceDoesNotExist",
	)
	require.Error(t, errValidateResourceExists)
	require.ErrorIs(t, errValidateResourceExists, ErrResourceDoesNotExistOnTargetPolicy)

	errValidateResourceExists = localACP.ValidateResourceExistsOnValidDPI(
		ctx,
		"invalidPolicyID",
		"resourceDoesNotExist",
	)
	require.Error(t, errValidateResourceExists)
	require.ErrorIs(t, errValidateResourceExists, ErrPolicyDoesNotExistWithACP)

	errClose = localACP.Close()
	require.Nil(t, errClose)
}

func Test_LocalACP_InMemory_IsDocRegistered_TrueIfRegisteredFalseIfNotAndErrorOtherwise(t *testing.T) {
	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, "")
	errStart := localACP.Start(ctx)
	require.Nil(t, errStart)

	policyID, errAddPolicy := localACP.AddPolicy(
		ctx,
		identity1,
		validPolicy,
	)
	require.Nil(t, errAddPolicy)
	require.Equal(
		t,
		validPolicyID,
		policyID,
	)

	// Invalid empty doc and empty resource can't be registered.
	errRegisterDoc := localACP.RegisterDocObject(
		ctx,
		identity1,
		validPolicyID,
		"",
		"",
	)
	require.Error(t, errRegisterDoc)
	require.ErrorIs(t, errRegisterDoc, ErrFailedToRegisterDocWithACP)

	// Check if an invalid empty doc and empty resource is registered.
	isDocRegistered, errDocRegistered := localACP.IsDocRegistered(
		ctx,
		validPolicyID,
		"",
		"",
	)
	require.Error(t, errDocRegistered)
	require.ErrorIs(t, errDocRegistered, ErrFailedToCheckIfDocIsRegisteredWithACP)
	require.False(t, isDocRegistered)

	// No documents are registered right now so return false.
	isDocRegistered, errDocRegistered = localACP.IsDocRegistered(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errDocRegistered)
	require.False(t, isDocRegistered)

	// Register a document.
	errRegisterDoc = localACP.RegisterDocObject(
		ctx,
		identity1,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errRegisterDoc)

	// Now it should be registered.
	isDocRegistered, errDocRegistered = localACP.IsDocRegistered(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)

	require.Nil(t, errDocRegistered)
	require.True(t, isDocRegistered)

	errClose := localACP.Close()
	require.Nil(t, errClose)
}

func Test_LocalACP_PersistentMemory_IsDocRegistered_TrueIfRegisteredFalseIfNotAndErrorOtherwise(t *testing.T) {
	acpPath := t.TempDir()
	require.NotEqual(t, "", acpPath)

	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, acpPath)
	errStart := localACP.Start(ctx)
	require.Nil(t, errStart)

	policyID, errAddPolicy := localACP.AddPolicy(
		ctx,
		identity1,
		validPolicy,
	)
	require.Nil(t, errAddPolicy)
	require.Equal(
		t,
		validPolicyID,
		policyID,
	)

	// Invalid empty doc and empty resource can't be registered.
	errRegisterDoc := localACP.RegisterDocObject(
		ctx,
		identity1,
		validPolicyID,
		"",
		"",
	)
	require.Error(t, errRegisterDoc)
	require.ErrorIs(t, errRegisterDoc, ErrFailedToRegisterDocWithACP)

	// Check if an invalid empty doc and empty resource is registered.
	isDocRegistered, errDocRegistered := localACP.IsDocRegistered(
		ctx,
		validPolicyID,
		"",
		"",
	)
	require.Error(t, errDocRegistered)
	require.ErrorIs(t, errDocRegistered, ErrFailedToCheckIfDocIsRegisteredWithACP)
	require.False(t, isDocRegistered)

	// No documents are registered right now so return false.
	isDocRegistered, errDocRegistered = localACP.IsDocRegistered(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errDocRegistered)
	require.False(t, isDocRegistered)

	// Register a document.
	errRegisterDoc = localACP.RegisterDocObject(
		ctx,
		identity1,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errRegisterDoc)

	// Now it should be registered.
	isDocRegistered, errDocRegistered = localACP.IsDocRegistered(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)

	require.Nil(t, errDocRegistered)
	require.True(t, isDocRegistered)

	// Should stay registered even after a restart.
	errClose := localACP.Close()
	require.Nil(t, errClose)

	localACP.Init(ctx, acpPath)
	errStart = localACP.Start(ctx)
	require.Nil(t, errStart)

	// Check after restart if it is still registered.
	isDocRegistered, errDocRegistered = localACP.IsDocRegistered(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)

	require.Nil(t, errDocRegistered)
	require.True(t, isDocRegistered)

	errClose = localACP.Close()
	require.Nil(t, errClose)
}

func Test_LocalACP_InMemory_CheckDocAccess_TrueIfHaveAccessFalseIfNotErrorOtherwise(t *testing.T) {
	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, "")
	errStart := localACP.Start(ctx)
	require.Nil(t, errStart)

	policyID, errAddPolicy := localACP.AddPolicy(
		ctx,
		identity1,
		validPolicy,
	)
	require.Nil(t, errAddPolicy)
	require.Equal(
		t,
		validPolicyID,
		policyID,
	)

	// Invalid empty arguments such that we can't check doc access.
	hasAccess, errCheckDocAccess := localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity1.DID,
		validPolicyID,
		"",
		"",
	)
	require.Error(t, errCheckDocAccess)
	require.ErrorIs(t, errCheckDocAccess, ErrFailedToVerifyDocAccessWithACP)
	require.False(t, hasAccess)

	// Check document accesss for a document that does not exist.
	hasAccess, errCheckDocAccess = localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity1.DID,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.False(t, hasAccess)

	// Register a document.
	errRegisterDoc := localACP.RegisterDocObject(
		ctx,
		identity1,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errRegisterDoc)

	// Now check using correct identity if it has access.
	hasAccess, errCheckDocAccess = localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity1.DID,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.True(t, hasAccess)

	// Now check using wrong identity, it should not have access.
	hasAccess, errCheckDocAccess = localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity2.DID,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.False(t, hasAccess)

	errClose := localACP.Close()
	require.Nil(t, errClose)
}

func Test_LocalACP_PersistentMemory_CheckDocAccess_TrueIfHaveAccessFalseIfNotErrorOtherwise(t *testing.T) {
	acpPath := t.TempDir()
	require.NotEqual(t, "", acpPath)

	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, acpPath)
	errStart := localACP.Start(ctx)
	require.Nil(t, errStart)

	policyID, errAddPolicy := localACP.AddPolicy(
		ctx,
		identity1,
		validPolicy,
	)
	require.Nil(t, errAddPolicy)
	require.Equal(
		t,
		validPolicyID,
		policyID,
	)

	// Invalid empty arguments such that we can't check doc access.
	hasAccess, errCheckDocAccess := localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity1.DID,
		validPolicyID,
		"",
		"",
	)
	require.Error(t, errCheckDocAccess)
	require.ErrorIs(t, errCheckDocAccess, ErrFailedToVerifyDocAccessWithACP)
	require.False(t, hasAccess)

	// Check document accesss for a document that does not exist.
	hasAccess, errCheckDocAccess = localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity1.DID,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.False(t, hasAccess)

	// Register a document.
	errRegisterDoc := localACP.RegisterDocObject(
		ctx,
		identity1,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errRegisterDoc)

	// Now check using correct identity if it has access.
	hasAccess, errCheckDocAccess = localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity1.DID,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.True(t, hasAccess)

	// Now check using wrong identity, it should not have access.
	hasAccess, errCheckDocAccess = localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity2.DID,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.False(t, hasAccess)

	// identities should continue having their correct behaviour and access even after a restart.
	errClose := localACP.Close()
	require.Nil(t, errClose)

	localACP.Init(ctx, acpPath)
	errStart = localACP.Start(ctx)
	require.Nil(t, errStart)

	// Now check again after the restart using correct identity if it still has access.
	hasAccess, errCheckDocAccess = localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity1.DID,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.True(t, hasAccess)

	// Now check again after restart using wrong identity, it should continue to not have access.
	hasAccess, errCheckDocAccess = localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity2.DID,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.False(t, hasAccess)

	errClose = localACP.Close()
	require.Nil(t, errClose)
}

func Test_LocalACP_InMemory_AddDocActorRelationship_FalseIfExistsBeforeTrueIfNoOp(t *testing.T) {
	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, "")
	errStart := localACP.Start(ctx)
	require.Nil(t, errStart)

	policyID, errAddPolicy := localACP.AddPolicy(
		ctx,
		identity1,
		validPolicy,
	)
	require.Nil(t, errAddPolicy)
	require.Equal(
		t,
		validPolicyID,
		policyID,
	)

	// Register a document.
	errRegisterDoc := localACP.RegisterDocObject(
		ctx,
		identity1,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errRegisterDoc)

	// Other identity does not have access yet.
	hasAccess, errCheckDocAccess := localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity2.DID,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.False(t, hasAccess)

	// Grant other identity access.
	exists, errAddDocActorRelationship := localACP.AddDocActorRelationship(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
		"reader",
		identity1,
		identity2.DID,
	)
	require.Nil(t, errAddDocActorRelationship)
	require.False(t, exists)

	// Granting again will be no-op
	exists, errAddDocActorRelationship = localACP.AddDocActorRelationship(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
		"reader",
		identity1,
		identity2.DID,
	)
	require.Nil(t, errAddDocActorRelationship)
	require.True(t, exists) // Exists already this time

	// Now the other identity has access.
	hasAccess, errCheckDocAccess = localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity2.DID,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.True(t, hasAccess)

	errClose := localACP.Close()
	require.Nil(t, errClose)
}

func Test_LocalACP_PersistentMemory_AddDocActorRelationship_FalseIfExistsBeforeTrueIfNoOp(t *testing.T) {
	acpPath := t.TempDir()
	require.NotEqual(t, "", acpPath)

	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, acpPath)
	errStart := localACP.Start(ctx)
	require.Nil(t, errStart)

	policyID, errAddPolicy := localACP.AddPolicy(
		ctx,
		identity1,
		validPolicy,
	)
	require.Nil(t, errAddPolicy)
	require.Equal(
		t,
		validPolicyID,
		policyID,
	)

	// Register a document.
	errRegisterDoc := localACP.RegisterDocObject(
		ctx,
		identity1,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errRegisterDoc)

	// Other identity does not have access yet.
	hasAccess, errCheckDocAccess := localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity2.DID,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.False(t, hasAccess)

	// Grant other identity access.
	exists, errAddDocActorRelationship := localACP.AddDocActorRelationship(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
		"reader",
		identity1,
		identity2.DID,
	)
	require.Nil(t, errAddDocActorRelationship)
	require.False(t, exists)

	// Granting again will be no-op
	exists, errAddDocActorRelationship = localACP.AddDocActorRelationship(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
		"reader",
		identity1,
		identity2.DID,
	)
	require.Nil(t, errAddDocActorRelationship)
	require.True(t, exists) // Exists already this time

	// Now the other identity has access.
	hasAccess, errCheckDocAccess = localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity2.DID,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.True(t, hasAccess)

	// Should continue having their correct behaviour and access even after a restart.
	errClose := localACP.Close()
	require.Nil(t, errClose)

	localACP.Init(ctx, acpPath)
	errStart = localACP.Start(ctx)
	require.Nil(t, errStart)

	// Now check again after the restart that the second identity still has access.
	hasAccess, errCheckDocAccess = localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity2.DID,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.True(t, hasAccess)

	errClose = localACP.Close()
	require.Nil(t, errClose)
}

func Test_LocalACP_InMemory_DeleteDocActorRelationship_TrueIfFoundAndDeletedFalseOtherwise(t *testing.T) {
	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, "")
	errStart := localACP.Start(ctx)
	require.Nil(t, errStart)

	policyID, errAddPolicy := localACP.AddPolicy(
		ctx,
		identity1,
		validPolicy,
	)
	require.Nil(t, errAddPolicy)
	require.Equal(
		t,
		validPolicyID,
		policyID,
	)

	// Register a document.
	errRegisterDoc := localACP.RegisterDocObject(
		ctx,
		identity1,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errRegisterDoc)

	// Grant other identity access.
	exists, errAddDocActorRelationship := localACP.AddDocActorRelationship(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
		"reader",
		identity1,
		identity2.DID,
	)
	require.Nil(t, errAddDocActorRelationship)
	require.False(t, exists)

	// Now the other identity has access.
	hasAccess, errCheckDocAccess := localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity2.DID,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.True(t, hasAccess)

	// Delete other identity's access by removing their relationship.
	foundRecord, errDeleteDocActorRelationship := localACP.DeleteDocActorRelationship(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
		"reader",
		identity1,
		identity2.DID,
	)
	require.Nil(t, errDeleteDocActorRelationship)
	require.True(t, foundRecord)

	// Deleting same relationship again should be no-op.
	foundRecord, errDeleteDocActorRelationship = localACP.DeleteDocActorRelationship(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
		"reader",
		identity1,
		identity2.DID,
	)
	require.Nil(t, errDeleteDocActorRelationship)
	require.False(t, foundRecord) // Is a no-op

	// Other identity now has no access again.
	hasAccess, errCheckDocAccess = localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity2.DID,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.False(t, hasAccess)

	errClose := localACP.Close()
	require.Nil(t, errClose)
}

func Test_LocalACP_PersistentMemory_DeleteDocActorRelationship_TrueIfFoundAndDeletedFalseOtherwise(t *testing.T) {
	acpPath := t.TempDir()
	require.NotEqual(t, "", acpPath)

	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, acpPath)
	errStart := localACP.Start(ctx)
	require.Nil(t, errStart)

	policyID, errAddPolicy := localACP.AddPolicy(
		ctx,
		identity1,
		validPolicy,
	)
	require.Nil(t, errAddPolicy)
	require.Equal(
		t,
		validPolicyID,
		policyID,
	)

	// Register a document.
	errRegisterDoc := localACP.RegisterDocObject(
		ctx,
		identity1,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errRegisterDoc)

	// Grant other identity access.
	exists, errAddDocActorRelationship := localACP.AddDocActorRelationship(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
		"reader",
		identity1,
		identity2.DID,
	)
	require.Nil(t, errAddDocActorRelationship)
	require.False(t, exists)

	// Now the other identity has access.
	hasAccess, errCheckDocAccess := localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity2.DID,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.True(t, hasAccess)

	// Delete other identity's access by removing their relationship.
	foundRecord, errDeleteDocActorRelationship := localACP.DeleteDocActorRelationship(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
		"reader",
		identity1,
		identity2.DID,
	)
	require.Nil(t, errDeleteDocActorRelationship)
	require.True(t, foundRecord)

	// Deleting same relationship again should be no-op.
	foundRecord, errDeleteDocActorRelationship = localACP.DeleteDocActorRelationship(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
		"reader",
		identity1,
		identity2.DID,
	)
	require.Nil(t, errDeleteDocActorRelationship)
	require.False(t, foundRecord) // Is a no-op

	// Other identity now has no access again.
	hasAccess, errCheckDocAccess = localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity2.DID,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.False(t, hasAccess)

	// Should continue having their correct behaviour and access even after a restart.
	errClose := localACP.Close()
	require.Nil(t, errClose)

	localACP.Init(ctx, acpPath)
	errStart = localACP.Start(ctx)
	require.Nil(t, errStart)

	// Now check again after the restart that the second identity still has no access.
	hasAccess, errCheckDocAccess = localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity2.DID,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.False(t, hasAccess)

	errClose = localACP.Close()
	require.Nil(t, errClose)
}

func Test_LocalACP_InMemory_AddPolicy_InvalidCreatorIDReturnsError(t *testing.T) {
	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, "")
	err := localACP.Start(ctx)
	require.Nil(t, err)

	policyID, err := localACP.AddPolicy(
		ctx,
		invalidIdentity,
		validPolicy,
	)

	require.ErrorIs(t, err, ErrInvalidActorID)
	require.Empty(t, policyID)

	err = localACP.Close()
	require.NoError(t, err)
}

func Test_LocalACP_Persistent_AddPolicy_InvalidCreatorIDReturnsError(t *testing.T) {
	acpPath := t.TempDir()
	require.NotEqual(t, "", acpPath)

	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, acpPath)
	err := localACP.Start(ctx)
	require.Nil(t, err)

	policyID, err := localACP.AddPolicy(
		ctx,
		invalidIdentity,
		validPolicy,
	)

	require.ErrorIs(t, err, ErrInvalidActorID)
	require.Empty(t, policyID)

	err = localACP.Close()
	require.NoError(t, err)
}

func Test_LocalACP_InMemory_RegisterObject_InvalidCreatorIDReturnsError(t *testing.T) {
	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, "")
	err := localACP.Start(ctx)
	require.Nil(t, err)

	err = localACP.RegisterDocObject(
		ctx,
		invalidIdentity,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)

	require.ErrorIs(t, err, ErrInvalidActorID)

	err = localACP.Close()
	require.NoError(t, err)
}

func Test_LocalACP_Persistent_RegisterObject_InvalidCreatorIDReturnsError(t *testing.T) {
	acpPath := t.TempDir()
	require.NotEqual(t, "", acpPath)

	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, acpPath)
	err := localACP.Start(ctx)
	require.Nil(t, err)

	err = localACP.RegisterDocObject(
		ctx,
		invalidIdentity,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)

	require.ErrorIs(t, err, ErrInvalidActorID)

	err = localACP.Close()
	require.NoError(t, err)
}

func Test_LocalACP_InMemory_AddDocActorRelationship_InvalidIdentitiesReturnError(t *testing.T) {
	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, "")
	err := localACP.Start(ctx)
	require.Nil(t, err)

	// Invalid requesting identity.
	exists, err := localACP.AddDocActorRelationship(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
		"reader",
		invalidIdentity,
		identity2.DID,
	)
	require.False(t, exists)
	require.ErrorIs(t, err, ErrInvalidActorID)

	// Invalid target actor.
	exists, err = localACP.AddDocActorRelationship(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
		"reader",
		identity1,
		invalidIdentity.DID,
	)
	require.False(t, exists)
	require.ErrorIs(t, err, ErrFailedToAddDocActorRelationshipWithACP)

	err = localACP.Close()
	require.NoError(t, err)
}

func Test_LocalACP_Persistent_AddDocActorRelationship_InvalidIdentitiesReturnError(t *testing.T) {
	acpPath := t.TempDir()
	require.NotEqual(t, "", acpPath)

	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, acpPath)
	err := localACP.Start(ctx)
	require.Nil(t, err)

	// Invalid requesting identity.
	exists, err := localACP.AddDocActorRelationship(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
		"reader",
		invalidIdentity,
		identity2.DID,
	)
	require.False(t, exists)
	require.ErrorIs(t, err, ErrInvalidActorID)

	// Invalid target actor.
	exists, err = localACP.AddDocActorRelationship(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
		"reader",
		identity1,
		invalidIdentity.DID,
	)
	require.False(t, exists)
	require.ErrorIs(t, err, ErrFailedToAddDocActorRelationshipWithACP)

	err = localACP.Close()
	require.NoError(t, err)
}

func Test_LocalACP_InMemory_DeleteDocActorRelationship_InvalidIdentitiesReturnError(t *testing.T) {
	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, "")
	err := localACP.Start(ctx)
	require.Nil(t, err)

	// Invalid requesting identity.
	exists, err := localACP.DeleteDocActorRelationship(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
		"reader",
		invalidIdentity,
		identity2.DID,
	)
	require.False(t, exists)
	require.ErrorIs(t, err, ErrInvalidActorID)

	// Invalid target actor.
	exists, err = localACP.DeleteDocActorRelationship(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
		"reader",
		identity1,
		invalidIdentity.DID,
	)
	require.False(t, exists)
	require.ErrorIs(t, err, ErrFailedToDeleteDocActorRelationshipWithACP)

	err = localACP.Close()
	require.NoError(t, err)
}

func Test_LocalACP_Persistent_DeleteDocActorRelationship_InvalidIdentitiesReturnError(t *testing.T) {
	acpPath := t.TempDir()
	require.NotEqual(t, "", acpPath)

	ctx := context.Background()
	localACP := NewLocalACP()

	localACP.Init(ctx, acpPath)
	err := localACP.Start(ctx)
	require.Nil(t, err)

	// Invalid requesting identity.
	exists, err := localACP.DeleteDocActorRelationship(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
		"reader",
		invalidIdentity,
		identity2.DID,
	)
	require.False(t, exists)
	require.ErrorIs(t, err, ErrInvalidActorID)

	// Invalid target actor.
	exists, err = localACP.DeleteDocActorRelationship(
		ctx,
		validPolicyID,
		"users",
		"documentID_XYZ",
		"reader",
		identity1,
		invalidIdentity.DID,
	)
	require.False(t, exists)
	require.ErrorIs(t, err, ErrFailedToDeleteDocActorRelationshipWithACP)

	err = localACP.Close()
	require.NoError(t, err)
}
