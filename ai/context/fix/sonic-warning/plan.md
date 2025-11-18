# Implementation Plan: Fix Sonic Warning

## Overview

Upgrade the `bytedance/sonic` dependency from v1.12.3 to v1.14.2 to gain Go 1.24+ support and eliminate the compatibility warning.

## Root Cause Analysis

The dependency chain:
```
defradb/keyring
  → cosmos-sdk/types
    → cosmossdk.io/log@v1.5.0
      → github.com/rs/zerolog (logging framework)
        → github.com/bytedance/sonic@v1.12.3 (fast JSON encoding)
```

**Why sonic is used**: `cosmossdk.io/log` uses zerolog for structured logging, and sonic is used internally for high-performance JSON marshaling when outputting JSON-formatted logs.

## Solution Approach

### Option 1: Upgrade Sonic (CHOSEN) ✅
- **Action**: Upgrade sonic from v1.12.3 to v1.14.2
- **Rationale**:
  - v1.14.2 adds official Go 1.24 and 1.25 support
  - Maintains backward compatibility with v1.12.3 API
  - Includes performance improvements and bug fixes
  - No breaking changes to cosmos-sdk integration
- **Risk**: Low - sonic is not exposed in cosmossdk.io/log's public API

### Option 2: Use Build Flag (REJECTED) ❌
- **Action**: Add `-checklinkname=0` linker flag to Makefile
- **Rationale**: Suppresses the Go 1.24 linkname check
- **Risk**: Treats symptom, not root cause; warning would persist at runtime

## Implementation Steps

### Phase 1: Dependency Upgrade
```bash
go get github.com/bytedance/sonic@v1.14.2
go mod tidy
```

**Expected Changes**:
- `go.mod`: sonic v1.12.3 → v1.14.2
- `go.sum`: Updated checksums
- Additional transitive dependencies may be upgraded

### Phase 2: Verification Testing

**Build Verification**:
```bash
make install
defradb version  # Should show no warnings
defradb --help   # Should show no warnings
```

**Integration Testing**:
```bash
go test ./keyring/...        # Tests cosmos-sdk keyring
go test ./acp/identity/...   # Tests crypto/identity with cosmos-sdk
go test ./crypto/...         # Tests cryptographic operations
go build ./...               # Ensures entire project compiles
```

**Functional Verification**:
- Test cosmos-sdk types initialization
- Verify JSON serialization compatibility
- Compare output with standard library JSON

### Phase 3: Compatibility Analysis

**Dependency Impact**:
- Check `cosmossdk.io/log@v1.5.0` expects `sonic@v1.12.3`
- Verify minor version upgrade (v1.12 → v1.14) maintains compatibility
- Confirm no breaking API changes in sonic release notes

**Test Coverage**:
- Keyring operations (uses cosmos-sdk heavily)
- ACP identity/crypto features
- CLI integration tests
- P2P replication (may have timing issues unrelated to sonic)

## Affected Components

### Direct Changes
- `go.mod` - sonic version bump
- `go.sum` - checksum updates

### Indirect Impact (No Code Changes)
- `keyring/` - Uses cosmos-sdk crypto/types
- `acp/identity/` - Uses cosmos-sdk identity features
- `crypto/` - Uses cosmos-sdk cryptographic primitives

### No Impact
- Application logic
- Database operations
- Query processing
- P2P networking (except for any JSON serialization)
- HTTP API endpoints

## Rollback Strategy

If issues arise:
```bash
go get github.com/bytedance/sonic@v1.12.3
go mod tidy
```

This immediately reverts to the previous version.

## Testing Strategy

### Unit Tests
- Run existing test suites that exercise cosmos-sdk integration
- Focus on keyring, identity, and crypto packages

### Integration Tests
- CLI commands that use cosmos-sdk features
- Full application startup/shutdown
- Keyring import/export operations

### Manual Testing
```bash
# Test 1: No warnings
defradb version
defradb start --help

# Test 2: Keyring operations
defradb keyring generate
defradb keyring list

# Test 3: Node identity
defradb identity new
```

## Success Criteria

- [x] Sonic upgraded to v1.14.2
- [x] No sonic warnings on any command
- [x] All integration tests pass
- [x] Full project builds successfully
- [x] cosmos-sdk JSON serialization works correctly
- [x] No performance regression

## Notes

- This fix was already implemented in commit `0c9c4e41` on the `fix/sonic-warning` branch
- The branch includes the upgrade: "Bump sonic to version 14.2 with Go 1.24 support"
- All verification has been completed successfully
