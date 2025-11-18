# Learnings

## Sonic Usage in Cosmos SDK

**Discovery**: The `bytedance/sonic` library is used by `cosmossdk.io/log` for high-performance JSON marshaling in structured logging output, not for any protocol-level data serialization.

**Implication**:
- Sonic version changes have minimal impact on DefraDB functionality
- The dependency is purely for performance optimization in logs
- No risk to consensus, state machine, or data persistence
- Compatibility concerns are limited to JSON encoding correctness

**Source**: Traced via `go mod graph` and pkg.go.dev documentation for cosmossdk.io/log@v1.5.0

---

## Go 1.24 Linkname Restrictions

**Discovery**: Go 1.24 introduced stricter `//go:linkname` restrictions, which sonic v1.12.3 violates for performance optimizations.

**Context**:
- Sonic uses low-level Go internals for performance (JIT, SIMD)
- Go 1.24 added `-checklinkname` check to prevent unsafe linkname usage
- Sonic v1.14.0+ was released specifically to support Go 1.24/1.25

**Implication**:
- Libraries using advanced performance techniques may need updates for Go 1.24+
- Build flags can suppress warnings but don't fix runtime issues
- Always check library release notes for Go version compatibility

---

## Dependency Chain Visibility

**Discovery**: Sonic is 3 levels deep in the dependency chain:
```
defradb → cosmos-sdk/types → cosmossdk.io/log → sonic
```

**Lesson**:
- Use `go mod why -m <package>` to understand dependency necessity
- Use `go mod graph` to visualize full dependency tree
- Indirect dependencies can still cause visible issues (warnings, errors)
- Go's module system handles transitive upgrades automatically

**Tool Usage**:
```bash
go mod why github.com/bytedance/sonic
go mod graph | grep sonic
go list -m -versions github.com/bytedance/sonic
```

---

## Semantic Versioning Guarantees

**Discovery**: Sonic maintains strict semantic versioning - minor version bumps (v1.12 → v1.14) guarantee backward compatibility.

**Verification**:
- API compatibility: Public interfaces unchanged
- Behavior compatibility: JSON output identical between versions
- Performance compatibility: Equal or improved performance

**Best Practice**: When upgrading indirect dependencies, verify the library follows semantic versioning and check their changelog for breaking changes.

---

## Test Strategy for Indirect Dependencies

**Discovery**: Testing packages that use cosmos-sdk indirectly validated sonic compatibility:
- `keyring/` - Heavy cosmos-sdk crypto usage
- `acp/identity/` - JWT signing with cosmos keys
- `crypto/` - Cryptographic operations

**Lesson**: Integration tests are more valuable than unit tests for validating indirect dependency upgrades, as they exercise real code paths through the dependency chain.

---

## Go Module Version Resolution

**Discovery**: When a dependency declares `require sonic@v1.12.3`, Go allows using a higher compatible version (v1.14.2) without the parent package needing to update their go.mod.

**Mechanism**: Minimum Version Selection (MVS)
- Dependencies declare minimum required versions
- Go allows using higher compatible versions (same major version)
- No need to wait for cosmossdk.io/log to update their go.mod

**Reference**: This is standard Go modules behavior as per `go help modules`
