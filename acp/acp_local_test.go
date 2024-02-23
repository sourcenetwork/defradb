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
)

var identity1 = "cosmos1zzg43wdrhmmk89z3pmejwete2kkd4a3vn7w969"
var identity2 = "cosmos1x25hhksxhu86r45hqwk28dd70qzux3262hdrll"

var validPolicyID string = "4f13c5084c3d0e1e5c5db702fceef84c3b6ab948949ca8e27fcaad3fb8bc39f4"
var validPolicy string = `
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
	var localACP ACPLocal

	localACP.Init(ctx, "")
	err := localACP.Start(ctx)

	require.Nil(t, err)

	err = localACP.Close()
	require.Nil(t, err)
}

func Test_LocalACP_PersistentMemory_StartAndClose_NoError(t *testing.T) {
	acpModulePath := t.TempDir()
	require.NotEqual(t, "", acpModulePath)

	ctx := context.Background()
	var localACP ACPLocal

	localACP.Init(ctx, acpModulePath)
	err := localACP.Start(ctx)
	require.Nil(t, err)

	err = localACP.Close()
	require.Nil(t, err)
}

func Test_LocalACP_InMemory_AddPolicy_CanCreateTwice(t *testing.T) {
	ctx := context.Background()
	var localACP ACPLocal

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

	// Since nothing is persisted should allow adding same policy again.

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

func Test_LocalACP_PersistentMemory_AddPolicy_CanNotCreateTwice(t *testing.T) {
	acpModulePath := t.TempDir()
	require.NotEqual(t, "", acpModulePath)

	ctx := context.Background()
	var localACP ACPLocal

	localACP.Init(ctx, acpModulePath)
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

	// The above policy should remain persisted on restarting ACP module.

	localACP.Init(ctx, acpModulePath)
	errStart = localACP.Start(ctx)
	require.Nil(t, errStart)

	// Should not allow us to create the same policy again as it exists already.
	_, errAddPolicy = localACP.AddPolicy(
		ctx,
		identity1,
		validPolicy,
	)
	require.Error(t, errAddPolicy)
	require.ErrorIs(t, errAddPolicy, ErrFailedToAddPolicyWithACP)

	errClose = localACP.Close()
	require.Nil(t, errClose)
}

func Test_LocalACP_InMemory_ValidateResourseExistsOrNot_ErrIfDoesntExist(t *testing.T) {
	ctx := context.Background()
	var localACP ACPLocal

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
	require.ErrorIs(t, errValidateResourceExists, ErrPolicyDoesNotExistOnACPModule)

	errClose := localACP.Close()
	require.Nil(t, errClose)
}

func Test_LocalACP_PersistentMemory_ValidateResourseExistsOrNot_ErrIfDoesntExist(t *testing.T) {
	acpModulePath := t.TempDir()
	require.NotEqual(t, "", acpModulePath)

	ctx := context.Background()
	var localACP ACPLocal

	localACP.Init(ctx, acpModulePath)
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

	localACP.Init(ctx, acpModulePath)
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
	require.ErrorIs(t, errValidateResourceExists, ErrPolicyDoesNotExistOnACPModule)

	errClose = localACP.Close()
	require.Nil(t, errClose)
}

func Test_LocalACP_InMemory_IsDocRegistered_TrueIfRegisteredFalseIfNotAndErrorOtherwise(t *testing.T) {
	ctx := context.Background()
	var localACP ACPLocal

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
	acpModulePath := t.TempDir()
	require.NotEqual(t, "", acpModulePath)

	ctx := context.Background()
	var localACP ACPLocal

	localACP.Init(ctx, acpModulePath)
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

	localACP.Init(ctx, acpModulePath)
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
	var localACP ACPLocal

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
		identity1,
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
		identity1,
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
		identity1,
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
		identity2,
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
	acpModulePath := t.TempDir()
	require.NotEqual(t, "", acpModulePath)

	ctx := context.Background()
	var localACP ACPLocal

	localACP.Init(ctx, acpModulePath)
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
		identity1,
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
		identity1,
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
		identity1,
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
		identity2,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.False(t, hasAccess)

	// identities should continue having their correct behaviour and access even after a restart.
	errClose := localACP.Close()
	require.Nil(t, errClose)

	localACP.Init(ctx, acpModulePath)
	errStart = localACP.Start(ctx)
	require.Nil(t, errStart)

	// Now check again after the restart using correct identity if it still has access.
	hasAccess, errCheckDocAccess = localACP.CheckDocAccess(
		ctx,
		ReadPermission,
		identity1,
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
		identity2,
		validPolicyID,
		"users",
		"documentID_XYZ",
	)
	require.Nil(t, errCheckDocAccess)
	require.False(t, hasAccess)

	errClose = localACP.Close()
	require.Nil(t, errClose)
}
