// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package acp_types

import (
	"strings"

	"github.com/sourcenetwork/acp_core/pkg/types"
)

// RequiredRegistererRelationName is the required relation name that any registerer will have,
// as the registerer is guaranteed to be the owner.
const RequiredRegistererRelationName string = "owner"

// ACPSystemType is an enum type that indicates the type of the ACP system.
type ACPSystemType int

const (
	LocalDocumentACP ACPSystemType = iota
	SourceHubDocumentACP
	AdminACP
)

var _ ResourceInterfacePermission = (*DocumentResourcePermission)(nil)
var _ ResourceInterfacePermission = (*AdminResourcePermission)(nil)

type ResourceInterfacePermission interface {
	String() string
}

// DocumentResourcePermission is a resource interface permission for document access control.
type DocumentResourcePermission int

// Resource interface permission types for document access control.
const (
	DocumentReadPerm DocumentResourcePermission = iota
	DocumentUpdatePerm
	DocumentDeletePerm
)

// List of all valid resource interface permissions for document access control, the order of
// permissions in this list must match the above defined ordering such that iota matches the
// index position within the list.
var RequiredResourcePermissionsForDocument = []string{
	"read",
	"update",
	"delete",
}

func (resourcePermission DocumentResourcePermission) String() string {
	return RequiredResourcePermissionsForDocument[resourcePermission]
}

// ImplyDocumentReadPerm is a list of permissions that imply user can read. This is because
// for DefraDB's document access control purposes if an identity has access to any write
// permission (delete or update), then they don't need to explicitly have read permission to
// read, we just imply that they have read access.
var ImplyDocumentReadPerm = []DocumentResourcePermission{
	DocumentReadPerm,
	DocumentUpdatePerm,
	DocumentDeletePerm,
}

// AdminResourcePermission is a resource interface permission for admin access control.
type AdminResourcePermission int

// Resource interface permission types for admin access control.
const (
	AdminDACBypassPerm AdminResourcePermission = iota
	AdminDACEnablePerm
	AdminDACDisablePerm
	AdminDACPurgePerm
	AdminDACStatusPerm
	AdminDACRelationAddPerm
	AdminDACRelationDeletePerm
	AdminDACPolicyAddPerm
	AdminAACReEnablePerm
	AdminAACDisablePerm
	AdminAACPurgePerm
	AdminAACStatusPerm
	AdminAACRelationAddPerm
	AdminAACRelationDeletePerm
	AdminSchemaAddPerm
	AdminSchemaPatchPerm
	AdminPatchCollectionPerm
)

// List of all valid resource interface permissions for admin access control, the order of
// permissions in this list must match the above defined ordering such that iota matches the
// index position within the list.
var RequiredResourcePermissionsForAdmin = []string{
	"dac-bypass",
	"dac-enable",
	"dac-disable",
	"dac-purge",
	"dac-status",
	"dac-relation-add",
	"dac-relation-delete",
	"dac-policy-add",
	"aac-re-enable",
	"aac-disable",
	"aac-purge",
	"aac-status",
	"aac-relation-add",
	"aac-relation-delete",
	"schema-add",
	"schema-patch",
	"patch-collection",
}

const NodeACPObject = "NodeObject"

const NodeACPPolicyResourceName = "node"

const NodeACPPolicy = `
name: Node ACP Policy
description: Node ACP Policy

actor:
  name: actor

resources:
  node:
    permissions:
      dac-bypass:
        expr: owner + admin
      dac-enable:
        expr: owner + admin
      dac-disable:
        expr: owner + admin
      dac-purge:
        expr: owner + admin
      dac-status:
        expr: owner + admin
      dac-relation-add:
        expr: owner + admin
      dac-relation-delete:
        expr: owner + admin
      dac-policy-add:
        expr: owner + admin
      aac-re-enable:
        expr: owner + admin
      aac-disable:
        expr: owner + admin
      aac-purge:
        expr: owner + admin
      aac-status:
        expr: owner + admin
      aac-relation-add:
        expr: owner + admin
      aac-relation-delete:
        expr: owner + admin
      schema-add:
        expr: owner + admin
      schema-patch:
        expr: owner + admin
      patch-collection:
        expr: owner + admin

    relations:
      owner:
        types:
          - actor
      admin:
        manages:
          - admin
        types:
          - actor
`

func (resourcePermission AdminResourcePermission) String() string {
	return RequiredResourcePermissionsForAdmin[resourcePermission]
}

func (resourcePermission AdminResourcePermission) IsForAACOperation() bool {
	permission := resourcePermission.String()
	if len(permission) >= 3 && strings.EqualFold(permission[:3], "aac") {
		return true
	}
	return false
}

// RegistrationResult is an enum type which indicates the result of a RegisterObject call to SourceHub / ACP Core
type RegistrationResult int32

const (
	// NoOp indicates no action was take. The operation failed or the Object already existed and was active
	RegistrationResult_NoOp RegistrationResult = 0
	// Registered indicates the Object was sucessfuly registered to the Actor.
	RegistrationResult_Registered RegistrationResult = 1
	// Unarchived indicates that a previously deleted Object is active again.
	// Only the original owners can Unarchive an object.
	RegistrationResult_Unarchived RegistrationResult = 2
)

// PolicyMarshalType represents the format in which a policy
// is marshaled as
type PolicyMarshalType int32

const (
	PolicyMarshalType_YAML PolicyMarshalType = 1
	PolicyMarshalType_JSON PolicyMarshalType = 2
)

// Policy is a data container carrying the necessary data
// to verify whether a Policy meets resource interface requirements
type Policy struct {
	ID        string
	Resources map[string]*Resource
}

// Resource is a data container carrying the necessary data
// to verify whether it meets resource interface requirements.
type Resource struct {
	Name        string
	Permissions map[string]*Permission
}

// Permission is a data container carrying the necessary data
// to verify whether it meets resource interface requirements.
type Permission struct {
	Name       string
	Expression string
}

func MapACPCorePolicy(pol *types.Policy) Policy {
	resources := make(map[string]*Resource)
	for _, coreResource := range pol.Resources {
		resource := MapACPCoreResource(coreResource)
		resources[resource.Name] = resource
	}

	return Policy{
		ID:        pol.Id,
		Resources: resources,
	}
}

func MapACPCoreResource(policy *types.Resource) *Resource {
	perms := make(map[string]*Permission)
	for _, corePermission := range policy.Permissions {
		perm := MapACPCorePermission(corePermission)
		perms[perm.Name] = perm
	}

	return &Resource{
		Name:        policy.Name,
		Permissions: perms,
	}
}

func MapACPCorePermission(perm *types.Permission) *Permission {
	return &Permission{
		Name:       perm.Name,
		Expression: perm.Expression,
	}
}
