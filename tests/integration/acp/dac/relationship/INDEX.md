# Index: `tests/integration/acp/dac/relationship`

## Overview

This directory contains integration tests for DAC (Document Access Control) relationship management operations. The subdirectories test the complete lifecycle of access relationships between documents and actors — both granting and revoking permissions — including input validation, authorization enforcement, delegation through managers, and edge cases such as public documents, wildcard actor targets, and collections without a policy.

## Subdirectories

| Directory | Summary |
|---|---|
| [`doc_actor/`](doc_actor/INDEX.md) | Tests for adding and deleting doc-actor relationships, covering owner and manager grants of read, update, and delete access, revocation of those grants, input validation, idempotency, self-revocation prohibition, wildcard actor targets, and policy-less collections. |
