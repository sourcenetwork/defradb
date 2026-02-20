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
	NodeACP
)

var _ ResourceInterfacePermission = (*DocumentResourcePermission)(nil)
var _ ResourceInterfacePermission = (*NodeResourcePermission)(nil)

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

// RequiredResourcePermissionsForDocument lists all valid resource interface permissions for
// document access control, the order of permissions in this list must match the above defined
// ordering such that iota matches the index position within the list.
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

// NodeResourcePermission is a resource interface permission for node access control.
type NodeResourcePermission int

// Resource interface permission types for node access control.
const (
	NodeDACBypassPerm NodeResourcePermission = iota
	NodeDACEnablePerm
	NodeDACDisablePerm
	NodeDACPurgePerm
	NodeDACStatusPerm
	NodeDACRelationAddPerm
	NodeDACRelationDeletePerm
	NodeDACPolicyAddPerm
	NodeNACReEnablePerm
	NodeNACDisablePerm
	NodeNACPurgePerm
	NodeNACStatusPerm
	NodeNACRelationAddPerm
	NodeNACRelationDeletePerm
	NodeCollectionPatchPerm
	NodeCollectionGetPerm
	NodeCollectionTruncatePerm
	NodeDocumentReadPerm
	NodeDocumentUpdatePerm
	NodeDocumentDeletePerm
	NodeIndexListPerm
	NodeIndexCreatePerm
	NodeIndexDeletePerm
	NodeEncryptedIndexAddPerm
	NodeEncryptedIndexDeletePerm
	NodeEncryptedIndexListPerm
	NodeEncryptedIndexListAllPerm
	NodeP2PPeerInfo
	NodeP2PPeerConnectPerm
	NodeP2PPeerActivePerm
	NodeP2PReplicatorAddPerm
	NodeP2PReplicatorDeletePerm
	NodeP2PReplicatorListPerm
	NodeP2PCollectionAddPerm
	NodeP2PCollectionDeletePerm
	NodeP2PCollectionListPerm
	NodeP2PDocumentAddPerm
	NodeP2PDocumentDeletePerm
	NodeP2PDocumentListPerm
	NodeP2PSyncDocumentsPerm
	NodeP2PSyncCollectionVersionsPerm
	NodeP2PSyncBranchableCollectionPerm
	NodeSignatureVerifyPerm
	NodeLensCreatePerm
	NodeLensListPerm
	NodeViewRefreshPerm
	NodeViewAddPerm
	NodeMigrationSetPerm
)

// RequiredResourcePermissionsForNode lists all valid resource interface permissions for
// node access control, the order of permissions in this list must match the above defined
// ordering such that iota matches the index position within the list.
var RequiredResourcePermissionsForNode = []string{
	"dac-bypass",
	"dac-enable",
	"dac-disable",
	"dac-purge",
	"dac-status",
	"dac-relation-add",
	"dac-relation-delete",
	"dac-policy-add",
	"nac-re-enable",
	"nac-disable",
	"nac-purge",
	"nac-status",
	"nac-relation-add",
	"nac-relation-delete",
	"collection-patch",
	"collection-get",
	"collection-truncate",
	"document-read",
	"document-update",
	"document-delete",
	"index-list",
	"index-create",
	"index-delete",
	"encrypted-index-add",
	"encrypted-index-delete",
	"encrypted-index-list",
	"encrypted-index-list-all",
	"p2p-peer-info",
	"p2p-peer-connect",
	"p2p-peer-active",
	"p2p-replicator-add",
	"p2p-replicator-delete",
	"p2p-replicator-list",
	"p2p-collection-add",
	"p2p-collection-delete",
	"p2p-collection-list",
	"p2p-document-add",
	"p2p-document-delete",
	"p2p-document-list",
	"p2p-sync-documents",
	"p2p-sync-collection-versions",
	"p2p-sync-branchable-collection",
	"signature-verify",
	"lens-create",
	"lens-list",
	"view-refresh",
	"view-add",
	"migration-set",
}

const NodeACPObject = "NodeObject"

const NodeACPPolicyResourceName = "node"

const NodeACPPolicy = `
description: Node ACP Policy
name: Node ACP Policy
resources:
- name: node
  permissions:
  - name: dac-bypass
    expr: admin
  - name: dac-enable
    expr: admin
  - name: dac-disable
    expr: admin
  - name: dac-purge
    expr: admin
  - name: dac-status
    expr: admin
  - name: dac-relation-add
    expr: admin
  - name: dac-relation-delete
    expr: admin
  - name: dac-policy-add
    expr: admin

  - name: nac-re-enable
    expr: admin
  - name: nac-disable
    expr: admin
  - name: nac-purge
    expr: admin
  - name: nac-status
    expr: admin
  - name: nac-relation-add
    expr: admin
  - name: nac-relation-delete
    expr: admin

  - name: collection-patch
    expr: admin
  - name: collection-get
    expr: admin
  - name: collection-truncate
    expr: admin

  - name: document-read
    expr: admin
  - name: document-update
    expr: admin
  - name: document-delete
    expr: admin

  - name: index-list
    expr: admin
  - name: index-create
    expr: admin
  - name: index-delete
    expr: admin

  - name: encrypted-index-add
    expr: admin
  - name: encrypted-index-delete
    expr: admin
  - name: encrypted-index-list
    expr: admin
  - name: encrypted-index-list-all
    expr: admin

  - name: p2p-peer-info
    expr: admin
  - name: p2p-peer-connect
    expr: admin
  - name: p2p-peer-active
    expr: admin
  - name: p2p-replicator-add
    expr: admin
  - name: p2p-replicator-delete
    expr: admin
  - name: p2p-replicator-list
    expr: admin
  - name: p2p-collection-add
    expr: admin
  - name: p2p-collection-delete
    expr: admin
  - name: p2p-collection-list
    expr: admin
  - name: p2p-document-add
    expr: admin
  - name: p2p-document-delete
    expr: admin
  - name: p2p-document-list
    expr: admin
  - name: p2p-sync-documents
    expr: admin
  - name: p2p-sync-collection-versions
    expr: admin
  - name: p2p-sync-branchable-collection
    expr: admin

  - name: signature-verify
    expr: admin

  - name: lens-create
    expr: admin
  - name: lens-list
    expr: admin

  - name: view-refresh
    expr: admin
  - name: view-add
    expr: admin

  - name: migration-set
    expr: admin

  relations:
  - name: admin
    manages:
    - admin
    types:
    - actor
`

func (resourcePermission NodeResourcePermission) String() string {
	return RequiredResourcePermissionsForNode[resourcePermission]
}

func (resourcePermission NodeResourcePermission) IsForNACOperation() bool {
	permission := resourcePermission.String()
	if len(permission) >= 3 && strings.EqualFold(permission[:3], "nac") {
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
	PolicyMarshalType_YAML PolicyMarshalType = PolicyMarshalType(types.PolicyMarshalingType_YAML)
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
