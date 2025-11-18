# SyncBranchableCollection Feature Documentation

## Quick Reference

This feature adds the ability to sync branchable collection DAGs from peer nodes via P2P.

**Implementation Status:** ✅ Complete (except C bindings header generation)

**Test Status:** ✅ 4/4 integration tests passing

## Files

- `specs.md` - Requirements and acceptance criteria
- `plan.md` - Implementation details and architecture
- `decisions.md` - Key decisions made during development
- `learnings.md` - Technical insights discovered
- `backlog.md` - Follow-up work (C bindings, documentation)
- `pull_request.md` - PR description ready for submission

## Summary

### What Changed

Added `SyncBranchableCollection(ctx, collectionName)` method across all client interfaces to enable syncing collection-level DAG from peers.

### Files Modified

**Core:**12 files
- Client interface, P2P implementation, DB wrappers, HTTP API, CLI

**Testing:** 5 files
- Test action, integration tests, client wrappers

**Generated:** 1 file (mocks)

### Usage Example

```bash
# CLI
defradb client p2p collection sync-branchable Users --timeout=10s

# Go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
err := db.SyncBranchableCollection(ctx, "Users")
```

### Test Results

```
PASS TestBranchableCollectionSync_WithSimpleBranchableCollection
PASS TestBranchableCollectionSync_WithMultipleDocuments
PASS TestBranchableCollectionSync_WithNonBranchableCollection_ShouldError
PASS TestBranchableCollectionSync_WithNonExistentCollection_ShouldError
```

## Next Steps

1. Review PR documentation in `pull_request.md`
2. Optional: Complete C bindings (see `backlog.md`)
3. Optional: Add user documentation (see `backlog.md`)
