# Index: `tests/integration/acp/dac/link_collection`

## Overview

This folder tests linking a collection to a DAC policy via the `@policy(id, resource)` schema directive. Tests verify that a collection is accepted only when it references a valid, existing policy whose named resource satisfies all Document Resource Interface (DRI) requirements — meaning the resource must declare `delete`, `read`, and `update` permissions. Collections referencing a non-existent policy, a missing or invalid resource name, absent required permissions, or malformed directive arguments are rejected and must not appear in the schema.

## Test Index

### `accept_basic_dri_fmts_test.go`

Verifies that a collection linked to a minimal, well-formed YAML DRI policy is accepted.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_LinkCollection_BasicYAML_SchemaAccepted` | 22-97 | Link collection to policy using a basic YAML DRI format and accept it. |

### `accept_extra_permissions_on_dri_test.go`

Verifies that extra permissions beyond the required DRI set do not prevent collection acceptance.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_LinkCollection_WithExtraPermsHavingRequiredRelation_AcceptCollection` | 22-105 | Link collection to policy with extra permissions beyond required DRI and accept it. |

### `accept_managed_relation_on_dri_test.go`

Verifies that a DRI resource with a managed relation is still accepted for collection linking.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_LinkCollection_WithManagedRelation_AcceptCollections` | 22-108 | Link collection to policy with a managed relation on the DRI and accept it. |

### `accept_mixed_resources_on_partial_dri_test.go`

Verifies that a collection linked to the valid resource in a partially-DRI-compliant policy is accepted even when another resource in the same policy is invalid.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_LinkCollection_UseValidResource_AcceptCollection` | 22-113 | Link collection to the valid resource in a partially-DRI-compliant policy and accept it. |

### `accept_multi_dris_test.go`

Verifies that the same policy structure added by two different actors can each back a separate accepted collection.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_LinkCollection_AddPolicyTwiceWithValidDRIByDifferentActorsAndUseBoth_AcceptCollection` | 23-168 | Link two collections to the same-resource policy added by different actors and accept both. |

### `accept_multi_resources_on_dri_test.go`

Verifies that a policy containing multiple resources can back one or both collections when each links to a valid resource.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_LinkCollection_WithMultipleResources_AcceptCollection` | 22-108 | Link collection to one valid resource in a multi-resource policy and accept it. |
| `TestACP_LinkCollection_WithMultipleResourcesBothBeingUsed_AcceptCollection` | 110-238 | Link two collections each to a different resource in the same policy and accept both. |

### `accept_permission_with_omitted_owner_authority_test.go`

Verifies that collections are accepted even when a permission expression references an actor relation that omits implicit owner authority, demonstrating ACP still enforces access correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_LinkCollection_MaliciousOwnerSpecifiedOnUpdatePermissionExprOnDRI_ACPEnforcesAccess` | 22-103 | Link collection where update permission expr omits owner authority and accept it. |
| `TestACP_LinkCollection_MaliciousOwnerSpecifiedOnReadPermissionExprOnDRI_ACPEnforcesOwnerAccess` | 105-185 | Link collection where read permission expr omits owner authority and accept it. |
| `TestACP_LinkCollection_MaliciousOwnerSpecifiedOnDeletePermissionExprOnDRI_ACPEnforcesOwnerAccess` | 187-268 | Link collection where delete permission expr omits owner authority and accept it. |

### `accept_same_resource_on_diff_collections_test.go`

Verifies that multiple collections can be linked to the same resource name within one policy.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_LinkCollection_UseSameResourceOnDifferentSchemas_AcceptCollections` | 23-161 | Link two different collections to the same resource on one policy and accept both. |

### `reject_empty_arg_on_collection_test.go`

Verifies that a collection is rejected when the `@policy` directive is missing arguments or provides empty string values for both `id` and `resource`.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_LinkCollection_NoArgWasSpecifiedOnCollection_CollectionRejected` | 21-81 | Reject collection linked to policy when no @policy arguments are provided. |
| `TestACP_LinkCollection_SpecifiedArgsAreEmptyOnCollection_CollectionRejected` | 83-144 | Reject collection linked to policy when both @policy id and resource arguments are empty strings. |

### `reject_invalid_arg_type_on_collection_test.go`

Verifies that a collection is rejected when the `@policy` directive receives a non-string (integer) value for the `id` or `resource` argument.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_LinkCollection_InvalidPolicyIDArgTypeWasSpecifiedOnCollection_CollectionRejected` | 21-81 | Reject collection when the @policy id argument is given a non-string integer type. |
| `TestACP_LinkCollection_InvalidResourceArgTypeWasSpecifiedOnCollection_CollectionRejected` | 83-144 | Reject collection when the @policy resource argument is given a non-string integer type. |

### `reject_invalid_owner_delete_perm_on_dri_test.go`

Verifies that a collection is rejected when its DRI resource is missing the required `delete` permission label.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_LinkCollection_OwnerMissingRequiredDeletePermissionLabelOnDRI_CollectionRejected` | 21-83 | Reject collection when the DRI resource is missing the required delete permission label. |

### `reject_invalid_owner_read_perm_on_dri_test.go`

Verifies that a collection is rejected when its DRI resource is missing the required `read` permission label.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_LinkCollection_OwnerMissingRequiredReadPermissionLabelOnDRI_CollectionRejected` | 21-83 | Reject collection when the DRI resource is missing the required read permission label. |

### `reject_invalid_owner_update_perm_on_dri_test.go`

Verifies that a collection is rejected when its DRI resource is missing the required `update` permission label.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_LinkCollection_OwnerMissingRequiredUpdatePermissionLabelOnDRI_CollectionRejected` | 21-83 | Reject collection when the DRI resource is missing the required update permission label. |

### `reject_missing_dri_test.go`

Verifies that a collection is rejected when the referenced policy ID does not exist in ACP, whether no policy was ever added or a different policy was added.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_LinkCollection_WhereNoPolicyWasAdded_CollectionRejected` | 22-69 | Reject collection referencing a policy ID that was never added to ACP. |
| `TestACP_LinkCollection_WhereAPolicyWasAddedButLinkedPolicyWasNotAdded_CollectionRejected` | 71-139 | Reject collection when the referenced policy ID does not match any added policy. |

### `reject_missing_id_arg_on_collection_test.go`

Verifies that a collection is rejected when the `@policy` directive omits the `id` argument or provides an empty string for it.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_LinkCollection_NoPolicyIDWasSpecifiedOnCollection_CollectionRejected` | 21-81 | Reject collection when the @policy directive omits the id argument entirely. |
| `TestACP_LinkCollection_SpecifiedPolicyIDArgIsEmptyOnCollection_CollectionRejected` | 83-143 | Reject collection when the @policy id argument is explicitly set to an empty string. |

### `reject_missing_perms_on_dri_test.go`

Verifies that a collection is rejected when its DRI resource defines no permissions at all.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_LinkCollection_MissingRequiredReadPermissionOnDRI_CollectionRejected` | 21-80 | Reject collection when the DRI resource has no permissions defined at all. |

### `reject_missing_resource_arg_on_collection_test.go`

Verifies that a collection is rejected when the `@policy` directive omits the `resource` argument or provides an empty string for it.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_LinkCollection_NoResourceWasSpecifiedOnCollection_CollectionRejected` | 21-81 | Reject collection when the @policy directive omits the resource argument entirely. |
| `TestACP_LinkCollection_SpecifiedResourceArgIsEmptyOnCollection_CollectionRejected` | 83-144 | Reject collection when the @policy resource argument is explicitly set to an empty string. |

### `reject_missing_resource_on_dri_test.go`

Verifies that a collection is rejected when the `resource` name supplied in `@policy` does not exist on the referenced policy.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_LinkCollection_SpecifiedResourceDoesNotExistOnDRI_CollectionRejected` | 21-85 | Reject collection when the specified resource name does not exist on the policy. |
