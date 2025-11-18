# Specifications: Fix Sonic Warning with Go 1.24

## Problem Statement

When running DefraDB built with Go 1.24.6, users encounter a warning message:
```
WARNING:(ast) sonic only supports go1.17~1.23, but your environment is not suitable
```

This warning appears because:
- DefraDB uses Go 1.24.6
- The `bytedance/sonic` JSON library (v1.12.3) only officially supports Go 1.17-1.23
- Sonic is an indirect dependency via `cosmossdk.io/log@v1.5.0`

## Requirements

### Functional Requirements
1. Eliminate the sonic warning when running DefraDB with Go 1.24+
2. Maintain full compatibility with existing cosmos-sdk integration
3. Ensure all existing tests continue to pass
4. No changes to DefraDB's public API or behavior

### Non-Functional Requirements
1. The solution must be minimal and focused
2. Must not introduce breaking changes to dependencies
3. Must be compatible with the existing cosmos-sdk v0.50.14
4. Should maintain or improve JSON serialization performance

## Acceptance Criteria

- [ ] `defradb version` runs without sonic warnings
- [ ] `defradb start` and all CLI commands run without warnings
- [ ] All integration tests pass (keyring, acp, crypto, cli)
- [ ] Full project builds without errors: `go build ./...`
- [ ] cosmos-sdk integration remains functional
- [ ] JSON serialization produces identical results
- [ ] No new dependencies added beyond sonic upgrade

## Constraints

- Must work with Go 1.24.6 (current version in use)
- Must maintain compatibility with cosmos-sdk v0.50.14
- Must not modify Makefile or build flags if sonic upgrade alone solves the issue
- Should follow semantic versioning principles for dependency upgrades

## Success Metrics

- Zero runtime warnings related to sonic
- 100% test pass rate maintained
- No performance regression in JSON operations
- Successful CI/CD pipeline execution
