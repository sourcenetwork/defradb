# Index: `tests/integration/acp`

## Overview

This directory contains the integration tests for DefraDB's Access Control Policy (ACP) system. ACP in DefraDB is composed of two complementary subsystems: DAC (Document Access Control) and NAC (Node Access Control). DAC controls who may read, update, or delete individual documents by associating each document with a policy stored in SourceHub or a local ACP backend; owners and delegated actors can share or revoke access at the per-document level. NAC controls which identities are permitted to perform node-level administrative and data operations at all — when NAC is enabled, every operation is gated behind an identity check, and only the node owner (and any identities they delegate via the NAC `admin` relation) may proceed. The two subsystems interact: the NAC `admin` relation grants an implicit DAC bypass, and the DAC system remains active regardless of NAC state. Together the subdirectories cover the full lifecycle of both systems, from policy registration and collection linking through document CRUD, aggregation, relation traversal, index management, P2P replication, and NAC lifecycle management.

## Subdirectories

| Directory | Summary |
|---|---|
| [`dac/`](dac/INDEX.md) | Tests the Document Access Control (DAC) system end-to-end, including policy registration, collection linking, per-document CRUD and aggregate access control, relation-traversal visibility, index-backed queries under DAC, P2P replication with DAC-protected collections, and doc-actor relationship management. |
| [`nac/`](nac/INDEX.md) | Tests the Node Access Control (NAC) system end-to-end, including NAC startup, enable/disable/re-enable lifecycle with restart persistence, access gates for every node operation (schema, document CRUD, P2P, index, lens, signature), NAC actor relationship management, DAC access patterns under NAC on/off states, and the delegated NAC admin relation. |
