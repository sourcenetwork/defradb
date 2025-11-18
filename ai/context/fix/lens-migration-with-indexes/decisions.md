# Decision Journal

This document records key decisions made during development when the initial approach didn't work or when alternatives were considered.

## Purpose

This file tracks:
- Failed attempts and why they failed
- Alternative approaches considered
- Rationale for chosen solutions
- Lessons learned during implementation

## Format

Each entry should include:
- Date
- Problem encountered
- Approaches tried
- Why they failed/succeeded
- Final decision and reasoning

---

## Entries

### 2025-11-12: Handling Branching Version Histories

**Problem**: PR review identified that `hasMigrationBetweenVersions` only checked linear chains, but collection versions can branch. The example given was:
```
       => C
A => B => D
```
If D is active and there's a migration from B to/from C, it should trigger reindexing.

**Initial Approach**: Use `lens.GetTargetedCollectionHistory()` to query the database for the full DAG.

**Why It Failed**: The lens history function queries fresh from the database using the collection version IDs. However, existing tests use hardcoded schema version ID constants (e.g., `schemaV1`, `schemaV2`) that don't match the actual version IDs of collections created by `PatchCollection`. This worked with the old linear chain approach because it used in-memory `colsWithRoot` parameter without validating against the DB.

**Final Decision**: Build a local DAG from the provided `colsWithRoot` parameter instead of querying the database. This approach:
1. Handles branching correctly by building a full DAG
2. Works with the existing test infrastructure that uses hardcoded version IDs
3. Doesn't require database queries (more efficient)
4. Maintains backward compatibility with existing tests

**Implementation**: Created local helper functions `buildCollectionHistoryDAG()` and `getTargetedHistory()` that mirror the lens package logic but work with in-memory collection lists.

**Note**: Existing tests still fail because they use incorrect hardcoded schema version IDs. These tests need to be fixed separately to use actual version IDs from created collections.

---

## Guidelines for Future Entries

When adding entries, consider documenting:
- Edge cases that required special handling
- Performance tradeoffs that were made
- Security considerations that influenced design
- Alternative implementations that were rejected
- Breaking changes that were avoided
- Compatibility concerns that were addressed
