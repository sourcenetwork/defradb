# Index: `tests/integration/acp/dac/add_policy`

## Overview

This folder tests the `AddDACPolicy` action in DefraDB's Data Access Control (DAC) system. The tests cover the full range of policy-upload behaviours: accepting valid DRI-compliant policies, accepting non-DRI-compliant-but-sourcehub-valid policies (e.g. no permissions, no resources), and rejecting malformed YAML, duplicate keys, invalid permission expressions, missing creator identity, and cross-resource relation references. Because policy upload is decoupled from collection registration, many structurally incomplete policies are accepted at this layer; DRI compliance is only enforced when a collection referencing the policy is added.

## Test Index

### `basic_test.go`

Tests that a minimal, DRI-compliant YAML policy is accepted and produces a valid policy ID.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddPolicy_BasicYAML_ValidPolicyID` | 20-43 | Add a valid DRI-compliant policy with basic YAML and receive a valid policy ID. |

### `with_empty_args_test.go`

Tests error handling when the policy data, the creator identity, or both are empty.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddPolicy_EmptyPolicyData_Error` | 20-36 | Adding a policy with empty policy data returns an error. |
| `TestACP_AddPolicy_EmptyPolicyCreator_Error` | 38-63 | Adding a policy with no identity (empty creator) returns an error. |
| `TestACP_AddPolicy_EmptyCreatorAndPolicyArgs_Error` | 65-81 | Adding a policy with both empty creator and empty policy data returns an error. |

### `with_empty_relations_test.go`

Tests that policies whose resources omit or leave blank the relations label are accepted.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddPolicy_NoRelationsLabel_NoError` | 20-41 | Add a policy with no relations label defined on the resource; succeeds. |
| `TestACP_AddPolicy_EmptyRelations_NoError` | 43-65 | Add a policy with an empty relations list on the resource; succeeds. |

### `with_empty_resource_test.go`

Tests that a policy with a single empty resource (no permissions, no relations) is accepted.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddPolicy_OneResourceThatIsEmpty_DoesntError` | 20-39 | Add a policy with a single resource that has no permissions or relations; succeeds. |

### `with_extra_perms_and_relations_test.go`

Tests that a policy combining extra non-DRI permissions with additional relations beyond the minimum is accepted.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddPolicy_ExtraPermissionsAndExtraRelations_ValidPolicyID` | 20-53 | Add a policy with extra non-DRI permissions and relations; returns a valid policy ID. |

### `with_extra_perms_test.go`

Tests policies with additional permissions beyond the DRI minimum, including the duplicate-key error case.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddPolicy_ExtraPermissions_ValidPolicyID` | 20-44 | Add a policy with extra non-DRI permissions alongside required ones; returns a valid policy ID. |
| `TestACP_AddPolicy_ExtraDuplicatePermissions_Error` | 46-85 | Add a policy with a duplicate permission key in YAML returns a parse error. |

### `with_extra_relations_test.go`

Tests policies with extra or duplicate relations, covering both the success and error paths.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddPolicy_ExtraRelations_ValidPolicyID` | 20-51 | Add a policy with extra unused relations beyond those referenced by permissions; succeeds. |
| `TestACP_AddPolicy_ExtraDuplicateRelations_Error` | 53-100 | Add a policy with a duplicate relation key in YAML returns a parse error. |

### `with_managed_relation_test.go`

Tests that a policy where one relation manages another relation is accepted.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddPolicy_WithRelationManagingOtherRelation_ValidPolicyID` | 20-53 | Add a policy where one relation manages another relation; returns a valid policy ID. |

### `with_multiple_resources_test.go`

Tests policies that declare multiple resources, including the error when one resource's permission references a relation from another resource.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddPolicy_MultipleResources_ValidID` | 20-58 | Add a policy with multiple resources each having their own permissions and relations; succeeds. |
| `TestACP_AddPolicy_MultipleResourcesUsingRelationDefinedInOther_Error` | 60-96 | Add a policy where one resource's permission references a relation defined only in another resource returns an error. |

### `with_multi_policies_test.go`

Tests scenarios where multiple policy-add operations are performed, covering different creators, identical policies, and format variations.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddPolicy_AddDuplicatePolicyByOtherCreator_ValidPolicyIDs` | 22-61 | Two different identities add the same policy and each receive a distinct valid policy ID. |
| `TestACP_AddPolicy_AddMultipleDuplicatePolicies_Error` | 63-109 | Same identity adding identical policies twice produces different expected policy IDs. |
| `TestACP_AddPolicy_AddMultipleDuplicatePoliciesDifferentFmts_ProducesDifferentIDs` | 111-157 | Adding identical policies in different YAML formats by the same creator produces different policy IDs. |
| `TestACP_AddPolicy_AddMultipleDifferentPolicies_ValidPolicyIDs` | 159-207 | Same identity adds two distinct policies sequentially and both receive valid policy IDs. |

### `with_no_perms_test.go`

Tests that non-DRI policies (missing permissions) are still accepted by the ACP layer, documenting sourcehub compatibility.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddPolicy_NoPermissionsOnlyOwner_ValidID` | 28-47 | Add a non-DRI policy with no permissions and only a resource name; returns a valid policy ID. |
| `TestACP_AddPolicy_NoPermissionsMultiRelations_ValidID` | 49-74 | Add a non-DRI policy with empty permissions label but multiple relations; returns a valid policy ID. |
| `TestACP_AddPolicy_NoPermissionsLabelSingleRelation_ValidID` | 76-97 | Add a non-DRI policy with no permissions label and a single relation; returns a valid policy ID. |
| `TestACP_AddPolicy_NoPermissionsLabelMultiRelations_ValidID` | 99-122 | Add a non-DRI policy with no permissions label and multiple typed relations; returns a valid policy ID. |

### `with_no_resources_test.go`

Tests that policies with no resources (incompatible with DRI but valid in sourcehub) are accepted, and that whitespace-only data is rejected.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddPolicy_NoResource_ValidID` | 25-43 | Add a non-DRI policy with an empty resources label; returns a valid policy ID. |
| `TestACP_AddPolicy_NoResourceLabel_ValidID` | 47-64 | Add a non-DRI policy with no resources label at all; returns a valid policy ID. |
| `TestACP_AddPolicy_PolicyWithOnlySpace_NameIsRequired` | 67-87 | Add a local-ACP policy whose data is only whitespace returns a name-required error. |

### `with_perm_expr_test.go`

Tests permission expression edge cases: the owner-in-expression error and the acceptance of permissions with no expression field.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddPolicy_PermissionExprWithOwnerInTheEndWithMinus_ErrorsBecauseOwnerIsInExpr` | 20-51 | Add a policy where a permission expression subtracts owner relation returns an error. |
| `TestACP_AddPolicy_EmptyExpressionInPermission_PermissionIsAccepted` | 53-80 | Add a policy with permissions that have no expression field; policy is accepted successfully. |

### `with_perm_invalid_expr_test.go`

Tests that invalid permission expressions (bad symbols, owner references, undeclared relations) are rejected.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddPolicy_PermissionExprWithIncorrectSymbol_Error` | 20-50 | Add a policy with an invalid symbol (^) in a permission expression returns a token recognition error. |
| `TestACP_AddPolicy_PermissionExprReferencingOwner_Error` | 52-80 | Add a policy where a permission expression directly references the owner relation returns an error. |
| `TestACP_AddPolicy_ExpressionReferencesUndeclaredRelation_Error` | 82-109 | Add a policy where a permission expression references a relation not declared in the policy returns an error. |

### `with_unused_relations_test.go`

Tests that a policy with a relation defined but not referenced by any permission expression is accepted.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddPolicy_UnusedRelation_ValidID` | 20-47 | Add a policy with a relation that is not referenced by any permission; returns a valid policy ID. |
