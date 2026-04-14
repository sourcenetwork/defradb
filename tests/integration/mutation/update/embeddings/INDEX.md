# Index: `tests/integration/mutation/update/embeddings`

## Overview

This folder contains integration tests for vector embedding behaviour during document update mutations. The tests verify that the `@embedding` directive correctly re-generates vector embeddings when source fields are updated, while ensuring generation is skipped when the vector field is supplied directly by the user or when none of the embedding source fields are part of the update.

## Test Index

### `embedding_test.go`

Tests cover the three key conditional paths for embedding generation on update: re-generation when source fields change, suppression when a vector is supplied explicitly, and suppression when only unrelated fields are updated.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationUpdate_WithMultipleEmbeddingFields_ShouldSucceed` | 26-103 | Update docs with multiple embedding fields re-generates vector embeddings. |
| `TestMutationUpdate_UserDefinedVectorEmbeddingDoesNotTriggerGeneration_ShouldSucceed` | 105-159 | Updating a vector field directly skips automatic embedding generation. |
| `TestMutationUpdate_FieldsForEmbeddingNotUpdatedDoesNotTriggerGeneration_ShouldSucceed` | 161-217 | Updating non-embedding fields does not trigger vector embedding re-generation. |
