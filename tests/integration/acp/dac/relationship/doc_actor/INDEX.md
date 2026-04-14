# Index: `tests/integration/acp/dac/relationship/doc_actor`

## Overview

This directory contains integration tests for DAC (Document Access Control) doc-actor relationship management. The subdirectories cover the full lifecycle of relationships between documents and actors: granting access (add) and revoking it (delete). Together they verify input validation, authorization rules, idempotency, delegation via managers, edge cases such as public documents and collections without a policy, and the correct propagation of read, update, and delete permissions — including GQL and non-GQL code paths.

## Subdirectories

| Directory | Summary |
|---|---|
| [`add/`](add/INDEX.md) | Tests for `AddDocActorRelationship`, covering valid grants of read, update, and delete access by owners and managers, input validation errors, idempotency, dummy relations, public documents, wildcard actor targets, and collections without a policy. |
| [`delete/`](delete/INDEX.md) | Tests for `DeleteDACActorRelationship`, covering revocation of reader, updater, deleter, and admin relationships, input validation errors, self-revocation prohibition, manager delegation revocation, wildcard relationship removal, and collections without a policy. |
