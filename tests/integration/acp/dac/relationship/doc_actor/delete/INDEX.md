# Index: `tests/integration/acp/dac/relationship/doc_actor/delete`

## Overview

This folder contains integration tests for the DAC (Document Access Control) doc-actor relationship deletion operation. The tests verify that the `DeleteDACActorRelationship` action correctly revokes previously granted relationships (reader, updater, deleter, admin/owner) between actors and documents, enforces the required input arguments, and upholds authorization rules such as preventing self-revocation and blocking operations on collections without a policy.

## Test Index

### `invalid_test.go`

Tests that `DeleteDACActorRelationship` returns appropriate errors when required arguments (docID, collection, relation name, target actor, or requestor identity) are missing.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_DeleteDocActorRelationshipMissingDocID_Error` | 21-107 | Deleting a doc-actor relationship without a docID returns an error. |
| `TestACP_DeleteDocActorRelationshipMissingCollection_Error` | 109-195 | Deleting a doc-actor relationship without a collection name returns an error. |
| `TestACP_DeleteDocActorRelationshipMissingRelationName_Error` | 197-283 | Deleting a doc-actor relationship without a relation name returns an error. |
| `TestACP_DeleteDocActorRelationshipMissingTargetActorName_Error` | 285-371 | Deleting a doc-actor relationship without a target actor identity returns an error. |
| `TestACP_DeleteDocActorRelationshipMissingReqestingIdentityName_Error` | 373-459 | Deleting a doc-actor relationship without a requestor identity returns an error. |

### `with_delete_test.go`

Tests that revoking a deleter relationship removes both delete and read access from the target actor.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_OwnerRevokesDeleteAccess_OtherActorCanNoLongerDelete` | 21-234 | Revoking the deleter relationship causes the actor to lose delete and read access. |

### `with_dummy_relation_test.go`

Tests the behavior of deletion requests that reference a relation name that is either defined on the policy but semantically inert, or entirely absent from the policy.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_DeleteDocActorRelationshipWithDummyRelationDefinedOnPolicy_NothingChanges` | 21-143 | Deleting a doc-actor relationship with a dummy relation defined on the policy has no effect. |
| `TestACP_DeleteDocActorRelationshipWithDummyRelationNotDefinedOnPolicy_Error` | 145-267 | Deleting a doc-actor relationship with a relation name not on the policy returns an error. |

### `with_manager_test.go`

Tests the delegation and revocation of management (admin) relationships, including manager-initiated revocations and forbidden attempts by an admin to revoke the owner.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_ManagerRevokesReadAccess_OtherActorCanNoLongerRead` | 21-175 | A manager (admin) revoking a reader relationship causes the actor to lose read access. |
| `TestACP_OwnerRevokesManagersAccess_ManagerCanNoLongerManageOthers` | 177-344 | Revoking the admin relationship prevents the manager from granting access to others. |
| `TestACP_AdminTriesToRevokeOwnersAccess_NotAllowedError` | 346-482 | An admin attempting to revoke the owner's relationship is forbidden. |

### `with_no_policy_on_collection_test.go`

Tests that attempting to delete a doc-actor relationship on a collection that has no ACP policy attached is rejected.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_DeleteDocActorRelationshipWithCollectionThatHasNoPolicy_NotAllowedError` | 21-65 | Deleting a doc-actor relationship on a collection with no ACP policy returns an error. |

### `with_public_document_test.go`

Tests that deleting a doc-actor relationship on a publicly accessible (ownerless) document is not allowed.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_DeleteDocActorRelationshipWithPublicDocument_CanAlreadyAccess_Error` | 21-129 | Deleting a doc-actor relationship on a public (ownerless) document returns an error. |

### `with_reader_test.go`

Tests that revoking a reader relationship removes read access, and that a second revocation of the same relationship is a no-op that reports the record was not found.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_OwnerRevokesReadAccessTwice_ShowThatTheRecordWasNotFoundSecondTime` | 21-135 | Revoking a reader relationship twice reports no record found on the second deletion. |
| `TestACP_OwnerRevokesGivenReadAccess_OtherActorCanNoLongerRead` | 137-279 | Revoking the reader relationship causes the actor to lose read access. |

### `with_self_test.go`

Tests that an actor (admin or owner) cannot revoke its own managed relationship.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_AdminTriesToRevokeItsOwnAccess_NotAllowedError` | 21-135 | An admin attempting to revoke its own admin relationship is forbidden. |
| `TestACP_OwnerTriesToRevokeItsOwnAccess_NotAllowedError` | 137-237 | An owner attempting to revoke its own owner relationship is forbidden. |

### `with_target_all_actors_test.go`

Tests that revoking a wildcard relationship (targeting all actors) removes implicit access from non-explicit actors while preserving explicit individual grants.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_OwnerRevokesAccessFromAllNonExplicitActors_ActorsCanNotReadAnymore` | 21-205 | Revoking the wildcard reader relationship removes read access from all non-explicit actors. |
| `TestACP_OwnerRevokesAccessFromAllNonExplicitActors_ExplicitActorsCanStillRead` | 207-515 | Revoking wildcard access only removes implicit access; explicitly granted actors can still read. |
| `TestACP_OwnerRevokesAccessFromAllNonExplicitActors_NonIdentityRequestsCanNotReadAnymore` | 517-659 | Revoking wildcard access removes read access even for unauthenticated (no-identity) requests. |

### `with_update_test.go`

Tests that revoking an updater relationship removes both update and read access from the target actor, covering both collection-mutation and GQL-mutation code paths.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_OwnerRevokesUpdateAccess_OtherActorCanNoLongerUpdate` | 24-224 | Revoking the updater relationship causes the actor to lose update and read access. |
| `TestACP_OwnerRevokesUpdateAccess_GQL_OtherActorCanNoLongerUpdate` | 226-426 | Revoking the updater relationship via GQL mutation causes the actor to lose update and read access. |
