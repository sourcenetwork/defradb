# Index: `tests/integration/mutation/add/embeddings`

## Overview

This folder contains integration tests for the automatic vector embedding generation feature in DefraDB. The tests verify that documents added to a collection with an `@embedding`-annotated field correctly trigger (or correctly skip) embedding generation via an external provider such as Ollama. They cover both the happy path — where embeddings are generated and stored as `Float32` vectors — and the bypass path where a caller supplies an explicit vector value that should be stored as-is without invoking the embedding provider.

## Test Index

### `embedding_test.go`

Tests that verify correct behaviour of the `@embedding` directive during document creation, including automatic generation and user-supplied vector bypass.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationAdd_WithMultipleEmbeddingFields_ShouldSucceed` | 26-89 | Adding docs with an embedding field generates vector embeddings for all documents. |
| `TestMutationAdd_UserDefinedVectorEmbeddingDoesNotTriggerGeneration_ShouldSucceed` | 91-133 | A user-supplied vector value on an embedding field skips automatic embedding generation. |
