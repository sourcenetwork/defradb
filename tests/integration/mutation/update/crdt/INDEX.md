# Index: `tests/integration/mutation/update/crdt`

## Overview

This folder contains integration tests that verify the behaviour of CRDT-typed fields during document mutation updates. It focuses on two counter CRDT types — PCounter (positive-only) and PNCounter (positive-and-negative) — across multiple numeric field kinds (Int, Float, Float32, Float64), asserting correct accumulation, rejection of invalid operations, and documented edge-case overflow behaviour.

## Test Index

### `pcounter_test.go`

Tests for the PCounter CRDT field type, covering increment validation, accumulation across numeric kinds, and overflow edge cases.

| Test Function | Line | Description |
|---|---|---|
| `TestPCounterUpdate_IntKindWithNegativeIncrement_ShouldError` | 27-75 | PCounter Int field rejects a negative increment with an error. |
| `TestPCounterUpdate_IntKindWithPositiveIncrement_ShouldIncrement` | 77-130 | PCounter Int field accumulates two positive increments correctly. |
| `TestPCounterUpdate_IntKindWithPositiveIncrementOverflow_RollsOverToMinInt64` | 133-194 | PCounter Int field rolls over to MinInt64 when Int64 maximum is exceeded. |
| `TestPCounterUpdate_FloatKindWithPositiveIncrement_ShouldIncrement` | 196-250 | PCounter Float field accumulates two positive increments correctly. |
| `TestPCounterUpdate_Float32KindWithPositiveIncrement_ShouldIncrement` | 252-305 | PCounter Float32 field accumulates two positive increments correctly. |
| `TestPCounterUpdate_Float64KindWithPositiveIncrement_ShouldIncrement` | 307-361 | PCounter Float64 field accumulates two positive increments correctly. |
| `TestPCounterUpdate_FloatKindWithPositiveIncrementOverflow_NoOp` | 365-412 | PCounter Float field increment beyond MaxFloat64 is a no-op. |

### `pncounter_test.go`

Tests for the PNCounter CRDT field type, covering positive and negative accumulation, overflow to infinity, and insignificant-value handling across numeric kinds.

| Test Function | Line | Description |
|---|---|---|
| `TestPNCounterUpdate_IntKindWithPositiveIncrement_ShouldIncrement` | 27-80 | PNCounter Int field accumulates two positive increments correctly. |
| `TestPNCounterUpdate_IntKindWithPositiveIncrementOverflow_RollsOverToMinInt64` | 83-144 | PNCounter Int field rolls over to MinInt64 when Int64 maximum is exceeded. |
| `TestPNCounterUpdate_FloatKindWithPositiveIncrement_ShouldIncrement` | 146-200 | PNCounter Float field accumulates two positive increments correctly. |
| `TestPNCounterUpdate_Float32KindWithPositiveIncrement_ShouldIncrement` | 202-255 | PNCounter Float32 field accumulates two positive increments correctly. |
| `TestPNCounterUpdate_Float64KindWithPositiveIncrement_ShouldIncrement` | 257-311 | PNCounter Float64 field accumulates two positive increments correctly. |
| `TestPNCounterUpdate_FloatKindWithPositiveIncrementOverflow_PositiveInf` | 314-368 | PNCounter Float field positive overflow beyond MaxFloat64 yields positive infinity. |
| `TestPNCounterUpdate_FloatKindWithDecrementOverflow_NegativeInf` | 371-425 | PNCounter Float field negative overflow beyond -MaxFloat64 yields negative infinity. |
| `TestPNCounterUpdate_FloatKindWithPositiveIncrementInsignificantValue_DoesNothing` | 427-475 | PNCounter Float field ignores an increment too small to affect a near-MaxFloat64 value. |
