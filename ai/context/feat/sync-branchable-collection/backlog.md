# Backlog

## ✅ C Bindings Completion (DONE)

**Task:** Complete C bindings for `SyncBranchableCollection`

**Status:** **COMPLETED** ✅

**What Was Done:**
1. Created `tools/scripts/build-c-shared-macos.sh` for macOS builds
2. Added `build-c-shared-macos` target to Makefile
3. Built C shared library: `build/libdefradb.dylib` (183 MB)
4. Generated C headers: `build/libdefradb.h` with `P2PbranchableCollectionSync` declaration
5. Implemented full C wrapper in `cbindings/wrapper.go`
6. Added extern declaration in cgo comment block
7. All 4 integration tests passing with `DEFRA_CLIENT_C=true`

**Files:**
- `tools/scripts/build-c-shared-macos.sh` (new)
- `Makefile` (added target)
- `cbindings/wrapper.go` (full implementation)
- `cbindings/p2p.go` (export function)
- `build/libdefradb.dylib` (generated)
- `build/libdefradb.h` (generated)

---

## JS Client Implementation

**Task:** Implement SyncBranchableCollection in JS/WASM client

**Current State:**
- Stubbed with `panic("not implemented")` in `tests/clients/js/wrapper.go`

**Note:** This is for test infrastructure. If DefraDB has a production JS client, that would need updating too.

---

## Enhanced Error Messages

**Potential Improvement:** Add more context to error messages

**Examples:**
- Include peer ID when no commits found
- Add collection ID alongside name in errors
- Include timeout duration in timeout errors

**Priority:** Low - current errors are functional

---

## Performance Optimization

**Observation:** Collection sync uses single-head assumption

**Current Code:**
```go
// Return the first (and should be only) head
return cids[0].Bytes(), nil
```

**Future Consideration:**
- If branchable collections ever support multiple heads (branching), this would need updating
- Could add validation that only one head exists
- Could handle multiple heads by syncing all

**Status:** Not needed now, collections have single head

---

## Documentation

**Needed:**
- User guide for `SyncBranchableCollection` use cases
- When to use vs. `SyncDocuments` vs. `FetchCollections`
- Examples of catch-up scenarios
- P2P networking guide updates

**Files:**
- README.md (P2P synchronization section)
- docs/website/guides/peer-to-peer.md
- CLI help text (already done)

**Priority:** Medium - helps user adoption
