# Index: `tests/integration/acp/dac/relationship/doc_actor/add`

## Overview

This folder contains integration tests for the DAC (Document Access Control) `AddDocActorRelationship` operation. Each test exercises the rules governing who may add a relationship between a document and an actor, and what capabilities that relationship subsequently grants. The tests cover valid owner and manager grants for read, update, and delete access, idempotency of duplicate relationship additions, error cases for missing or invalid arguments, and edge cases such as public documents and collections without a policy.

## Test Index

### `invalid_test.go`

Tests that verify the correct errors are returned when required arguments (doc ID, collection, relation name, target actor, or requesting identity) are missing from an `AddDocActorRelationship` call.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddDocActorRelationshipMissingDocID_Error` | 21-107 | Adding a doc-actor relationship fails when the doc ID is missing. |
| `TestACP_AddDocActorRelationshipMissingCollection_Error` | 109-195 | Adding a doc-actor relationship fails when the collection is missing. |
| `TestACP_AddDocActorRelationshipMissingRelationName_Error` | 197-283 | Adding a doc-actor relationship fails when the relation name is missing. |
| `TestACP_AddDocActorRelationshipMissingTargetActorName_Error` | 285-371 | Adding a doc-actor relationship fails when the target actor identity is missing. |
| `TestACP_AddDocActorRelationshipMissingReqestingIdentityName_Error` | 373-459 | Adding a doc-actor relationship fails when the requesting identity is missing. |

### `with_delete_test.go`

Tests that verify an owner can grant delete access to another actor, that the granted actor can then delete, and that duplicate grants are detected as already existing.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_OwnerGivesDeleteAccessToAnotherActorTwice_ShowThatTheRelationshipAlreadyExists` | 21-148 | Owner adding a delete relationship twice shows the relationship already exists. |
| `TestACP_OwnerGivesDeleteAccessToAnotherActor_OtherActorCanDelete` | 150-313 | Owner grants delete access to another actor, who can then delete the document. |
| `TestACP_OwnerGivesDeleteAccessToAnotherActor_OtherActorCanDeleteSoCanTheOwner` | 315-451 | Owner grants delete access to another actor; both the actor and owner can delete. |

### `with_dummy_relation_test.go`

Tests that verify the behavior when a relationship is added using a relation that has no effective permissions, whether the relation is defined on the policy or entirely absent.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddDocActorRelationshipWithDummyRelationDefinedOnPolicy_NothingChanges` | 21-143 | Adding a doc-actor relationship with a dummy relation on the policy grants no access. |
| `TestACP_AddDocActorRelationshipWithDummyRelationNotDefinedOnPolicy_Error` | 145-267 | Adding a doc-actor relationship with a relation not defined on the policy errors. |

### `with_manager_gql_test.go`

Tests for manager delegation scenarios using GQL mutations, verifying that an admin actor can self-assign roles and grant roles to others, but only for relations it manages.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_OwnerMakesAManagerThatGivesItSelfReadAndWriteAccess_GQL_ManagerCanReadAndWrite` | 24-279 | Owner makes a manager who self-assigns read and write, confirming full access via GQL. |
| `TestACP_OwnerMakesManagerButManagerCanNotPerformOperations_GQL_ManagerCantReadOrWrite` | 281-433 | Owner makes a manager who has no direct doc permissions and cannot read or write via GQL. |
| `TestACP_ManagerAddsRelationshipWithRelationItDoesNotManageAccordingToPolicy_GQL_Error` | 435-601 | Manager attempting to add a relationship for a relation it does not manage errors via GQL. |

### `with_manager_test.go`

Tests for manager delegation scenarios using non-GQL mutations, covering managers granting read/write to others, self-assigning roles, and unauthorized delegation attempts.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_ManagerGivesReadAccessToAnotherActor_OtherActorCanRead` | 24-192 | A manager grants read access to another actor, who can then read the document. |
| `TestACP_ManagerGivesWriteAccessToAnotherActor_OtherActorCanWrite` | 194-391 | A manager grants write access to another actor, who can then update and delete the document. |
| `TestACP_OwnerMakesAManagerThatGivesItSelfReadAccess_ManagerCanRead` | 393-561 | Owner makes a manager who self-assigns read access and can then read the document. |
| `TestACP_OwnerMakesAManagerThatGivesItSelfReadAndWriteAccess_ManagerCanReadAndWrite` | 563-819 | Owner makes a manager who self-assigns full access and can read, update, and delete. |
| `TestACP_ManagerAddsRelationshipWithRelationItDoesNotManageAccordingToPolicy_Error` | 821-986 | Manager attempting to add a relationship for a relation it does not manage returns an error. |
| `TestACP_OwnerMakesManagerButManagerCanNotPerformOperations_ManagerCantReadOrWrite` | 988-1139 | Owner makes a manager who has no direct doc permissions and cannot read or write. |
| `TestACP_CantMakeRelationshipIfNotOwnerOrManager_Error` | 1141-1229 | An actor that is neither owner nor manager cannot add a doc-actor relationship. |

### `with_no_policy_on_collection_test.go`

Tests that adding a doc-actor relationship on a collection without any ACP policy attached returns an error.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddDocActorRelationshipWithCollectionThatHasNoPolicy_NotAllowedError` | 21-65 | Adding a doc-actor relationship on a collection without a policy returns an error. |

### `with_only_delete_test.go`

Tests that an actor granted only the `deleter` relation can delete and implicitly read a document even when no separate read permission is configured on the policy.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_OwnerGivesDeleteAccessToAnotherActorWithoutExplicitReadPerm_OtherActorCanDelete` | 21-185 | Owner grants delete-only access; the actor can read and delete despite no explicit read permission. |

### `with_only_update_gql_test.go`

Tests via GQL that an actor granted only the `updater` relation can update and implicitly read without an explicit read permission in the policy.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_OwnerGivesUpdateAccessToAnotherActorWithoutExplicitReadPerm_GQL_OtherActorCanUpdate` | 24-188 | Owner grants update-only access via GQL; actor can update and implicitly read without explicit read permission. |

### `with_only_update_test.go`

Tests via non-GQL mutations that an actor granted only the `updater` relation can update and implicitly read without an explicit read permission in the policy.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_OwnerGivesUpdateAccessToAnotherActorWithoutExplicitReadPerm_OtherActorCanUpdate` | 24-189 | Owner grants update-only access; actor can update and implicitly read without an explicit read permission. |

### `with_public_document_test.go`

Tests that attempting to add a doc-actor relationship on a public document (created without an owner identity) errors because the document already has universal access.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AddDocActorRelationshipWithPublicDocument_CanAlreadyAccess_Error` | 21-129 | Adding a doc-actor relationship on a public document that is already accessible errors. |

### `with_reader_gql_test.go`

Tests via GQL that an actor granted only the `reader` relation can read but cannot update a document.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_OwnerGivesOnlyReadAccessToAnotherActor_GQL_OtherActorCanReadButNotUpdate` | 24-190 | Owner grants read-only access via GQL; actor can read but cannot update the document. |

### `with_reader_test.go`

Tests via non-GQL mutations covering owner granting read access to another actor, idempotent duplicate grants, ownership retention, and read-only access boundaries.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_OwnerGivesReadAccessToAnotherActorTwice_ShowThatTheRelationshipAlreadyExists` | 24-142 | Owner adding a read relationship twice shows the relationship already exists as a no-op. |
| `TestACP_OwnerGivesReadAccessToAnotherActor_OtherActorCanRead` | 144-272 | Owner grants read access to another actor, who can then read the document. |
| `TestACP_OwnerGivesReadAccessToAnotherActor_OtherActorCanReadSoCanTheOwner` | 276-410 | Owner grants read access to another actor; both the actor and owner retain read access. |
| `TestACP_OwnerGivesOnlyReadAccessToAnotherActor_OtherActorCanReadButNotUpdate` | 412-578 | Owner grants read-only access; actor can read the document but cannot update it. |
| `TestACP_OwnerGivesOnlyReadAccessToAnotherActor_OtherActorCanReadButNotDelete` | 580-728 | Owner grants read-only access; actor can read the document but cannot delete it. |

### `with_target_all_actors_gql_test.go`

Tests via GQL that targeting all actors with a read relationship allows any identity (including anonymous) to read but not mutate the document.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_OwnerGivesOnlyReadAccessToAllActors_GQL_AllActorsCanReadButNotUpdateOrDelete` | 24-232 | Owner grants read access to all actors via GQL; any actor can read but not update or delete. |
| `TestACP_OwnerGivesOnlyReadAccessToAllActors_GQL_CanReadEvenWithoutIdentityButNotUpdateOrDelete` | 234-420 | Owner grants read access to all actors via GQL; even an anonymous identity can read but not mutate. |

### `with_target_all_actors_test.go`

Tests via non-GQL mutations that targeting all actors with a read relationship allows any identity (including anonymous) to read but not mutate the document.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_OwnerGivesOnlyReadAccessToAllActors_AllActorsCanReadButNotUpdateOrDelete` | 24-236 | Owner grants read access to all actors; any actor can read but cannot update or delete. |
| `TestACP_OwnerGivesOnlyReadAccessToAllActors_CanReadEvenWithoutIdentityButNotUpdateOrDelete` | 238-426 | Owner grants read access to all actors; even an anonymous identity can read but not mutate. |

### `with_update_gql_test.go`

Tests via GQL that an owner can grant update access to another actor, that the actor can subsequently update the document, and that duplicate grants are idempotent.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_OwnerGivesUpdateAccessToAnotherActorTwice_GQL_ShowThatTheRelationshipAlreadyExists` | 24-164 | Owner adding an update relationship twice via GQL shows the relationship already exists. |
| `TestACP_OwnerGivesUpdateAccessToAnotherActor_GQL_OtherActorCanUpdate` | 166-328 | Owner grants update access to another actor via GQL; that actor can then update the document. |

### `with_update_test.go`

Tests via non-GQL mutations that an owner can grant update access to another actor, that both the actor and owner can update after the grant, and that duplicate grants are idempotent.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_OwnerGivesUpdateAccessToAnotherActorTwice_ShowThatTheRelationshipAlreadyExists` | 24-165 | Owner adding an update relationship twice shows the relationship already exists as a no-op. |
| `TestACP_OwnerGivesUpdateAccessToAnotherActor_OtherActorCanUpdate` | 167-332 | Owner grants update access to another actor, who can then update the document. |
| `TestACP_OwnerGivesUpdateAccessToAnotherActor_OtherActorCanUpdateSoCanTheOwner` | 334-496 | Owner grants update access to another actor; both the actor and owner retain update access. |
