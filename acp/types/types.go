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
	NodeBypassDACPerm NodeResourcePermission = iota
	NodeEnableDACPerm
	NodeDisableDACPerm
	NodePurgeDACPerm
	NodeGetDACStatusPerm
	NodeAddDACRelationPerm
	NodeDeleteDACRelationPerm
	NodeAddDACPolicyPerm
	NodeReEnableNACPerm
	NodeDisableNACPerm
	NodePurgeNACPerm
	NodeGetNACStatusPerm
	NodeAddNACRelationPerm
	NodeDeleteNACRelationPerm
	NodePatchCollectionPerm
	NodeGetCollectionPerm
	NodeTruncateCollectionPerm
	NodeReadDocumentPerm
	NodeUpdateDocumentPerm
	NodeDeleteDocumentPerm
	NodeListIndexPerm
	NodeAddIndexPerm
	NodeDeleteIndexPerm
	NodeAddEncryptedIndexPerm
	NodeDeleteEncryptedIndexPerm
	NodeListEncryptedIndexPerm
	NodeListAllEncryptedIndexPerm
	NodeConnectP2PPeerPerm
	NodeGetP2PPeerInfoPerm
	NodeGetP2PActivePeersPerm
	NodeAddP2PReplicatorPerm
	NodeDeleteP2PReplicatorPerm
	NodeListP2PReplicatorPerm
	NodeAddP2PCollectionPerm
	NodeDeleteP2PCollectionPerm
	NodeListP2PCollectionPerm
	NodeAddP2PDocumentPerm
	NodeDeleteP2PDocumentPerm
	NodeListP2PDocumentPerm
	NodeSyncP2PDocumentsPerm
	NodeSyncP2PCollectionVersionsPerm
	NodeSyncP2PBranchableCollectionPerm
	NodeVerifySignaturePerm
	NodeAddLensPerm
	NodeListLensPerm
	NodeRefreshViewPerm
	NodeAddViewPerm
	NodeSetMigrationPerm
)

// RequiredResourcePermissionsForNode lists all valid resource interface permissions for
// node access control, the order of permissions in this list must match the above defined
// ordering such that iota matches the index position within the list.
var RequiredResourcePermissionsForNode = []string{
	"bypass-dac",
	"enable-dac",
	"disable-dac",
	"purge-dac",
	"get-dac-status",
	"add-dac-relation",
	"delete-dac-relation",
	"add-dac-policy",
	"re-enable-nac",
	"disable-nac",
	"purge-nac",
	"get-nac-status",
	"add-nac-relation",
	"delete-nac-relation",
	"patch-collection",
	"get-collection",
	"truncate-collection",
	"read-document",
	"update-document",
	"delete-document",
	"list-index",
	"add-index",
	"delete-index",
	"add-encrypted-index",
	"delete-encrypted-index",
	"list-encrypted-index",
	"list-all-encrypted-index",
	"connect-p2p-peer",
	"get-p2p-peer-info",
	"get-p2p-active-peers",
	"add-p2p-replicator",
	"delete-p2p-replicator",
	"list-p2p-replicator",
	"add-p2p-collection",
	"delete-p2p-collection",
	"list-p2p-collection",
	"add-p2p-document",
	"delete-p2p-document",
	"list-p2p-document",
	"sync-p2p-documents",
	"sync-p2p-collection-versions",
	"sync-p2p-branchable-collection",
	"verify-signature",
	"add-lens",
	"list-lens",
	"refresh-view",
	"add-view",
	"set-migration",
}

const NodeACPObject = "NodeObject"

const NodeACPPolicyResourceName = "node"

const NodeACPPolicy = `
description: Node ACP Policy
name: Node ACP Policy
resources:
- name: node
  permissions:
  - name: bypass-dac
    expr: admin
  - name: enable-dac
    expr: admin
  - name: disable-dac
    expr: admin
  - name: purge-dac
    expr: admin
  - name: get-dac-status
    expr: admin
  - name: add-dac-relation
    expr: admin
  - name: delete-dac-relation
    expr: admin
  - name: add-dac-policy
    expr: admin

  - name: re-enable-nac
    expr: admin
  - name: disable-nac
    expr: admin
  - name: purge-nac
    expr: admin
  - name: get-nac-status
    expr: admin
  - name: add-nac-relation
    expr: admin
  - name: delete-nac-relation
    expr: admin

  - name: patch-collection
    expr: admin
  - name: get-collection
    expr: admin
  - name: truncate-collection
    expr: admin

  - name: read-document
    expr: admin
  - name: update-document
    expr: admin
  - name: delete-document
    expr: admin

  - name: list-index
    expr: admin
  - name: add-index
    expr: admin
  - name: delete-index
    expr: admin

  - name: add-encrypted-index
    expr: admin
  - name: delete-encrypted-index
    expr: admin
  - name: list-encrypted-index
    expr: admin
  - name: list-all-encrypted-index
    expr: admin

  - name: connect-p2p-peer
    expr: admin
  - name: get-p2p-peer-info
    expr: admin
  - name: get-p2p-active-peers
    expr: admin
  - name: add-p2p-replicator
    expr: admin
  - name: delete-p2p-replicator
    expr: admin
  - name: list-p2p-replicator
    expr: admin
  - name: add-p2p-collection
    expr: admin
  - name: delete-p2p-collection
    expr: admin
  - name: list-p2p-collection
    expr: admin
  - name: add-p2p-document
    expr: admin
  - name: delete-p2p-document
    expr: admin
  - name: list-p2p-document
    expr: admin
  - name: sync-p2p-documents
    expr: admin
  - name: sync-p2p-collection-versions
    expr: admin
  - name: sync-p2p-branchable-collection
    expr: admin

  - name: verify-signature
    expr: admin

  - name: add-lens
    expr: admin
  - name: list-lens
    expr: admin

  - name: refresh-view
    expr: admin
  - name: add-view
    expr: admin

  - name: set-migration
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
	switch resourcePermission {
	case NodeReEnableNACPerm,
		NodeDisableNACPerm,
		NodePurgeNACPerm,
		NodeGetNACStatusPerm,
		NodeAddNACRelationPerm,
		NodeDeleteNACRelationPerm:
		return true
	default:
		return false
	}
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
