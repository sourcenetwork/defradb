# Index: `tests/integration/node`

## Overview

This folder contains integration tests for node-level operations in DefraDB. Currently the tests cover node identity, verifying that each node in a multi-node setup correctly exposes its own unique cryptographic identity and does not return another node's identity.

## Test Index

### `identity_test.go`

Tests that each node correctly reports its own unique identity when queried.

| Test Function | Line | Description |
|---|---|---|
| `TestNodeIdentity_NodeIdentity_Succeed` | 20-38 | Each node returns its own unique identity when queried. |
