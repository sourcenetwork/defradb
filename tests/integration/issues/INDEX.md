# Index: `tests/integration/issues`

## Overview

This directory contains regression tests that guard against specific bugs filed as GitHub issues against DefraDB. Each test file is named after its corresponding issue number (e.g. `2566_test.go` covers issue #2566). The tests are intentionally kept even after the underlying bugs are fixed so that the same regressions cannot reappear silently.

## Test Index

### `2566_test.go`

Regression tests for issue #2566: simultaneous P2P pncounter float overflows on separate nodes produce inconsistent (non-converging) CRDT state.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PUpdate_WithPNCounterSimultaneousOverflowIncrement_DoesNotReachConsitency` | 28-137 | Simultaneous pncounter increment overflow on two nodes produces inconsistent float values. |
| `TestP2PUpdate_WithPNCounterSimultaneousOverflowDecrement_DoesNotReachConsitency` | 140-249 | Simultaneous pncounter decrement overflow on two nodes produces inconsistent float values. |

### `2569_test.go`

Regression tests for issue #2569: pncounter float overflow to ±infinity causes JSON serialization failures that prevent querying and collection access via the HTTP and CLI clients.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PUpdate_WithPNCounterFloatOverflowIncrement_PreventsQuerying` | 29-76 | Float pncounter increment overflow produces infinity, breaking HTTP/CLI query JSON parsing. |
| `TestP2PUpdate_WithPNCounterFloatOverflowDecrement_PreventsQuerying` | 78-125 | Float pncounter decrement overflow produces negative infinity, breaking HTTP/CLI query JSON parsing. |
| `TestP2PUpdate_WithPNCounterFloatOverflow_PreventsCollectionGet` | 127-183 | Float pncounter overflow to infinity causes collection.Get to return empty string, blocking subsequent updates. |
