# Decision Journal

## Decision 1: Upgrade Sonic vs. Use Linker Flag

**Date**: 2025-11-14

**Context**:
Two approaches were considered to fix the sonic warning:
1. Upgrade sonic to v1.14.2 (which supports Go 1.24+)
2. Add `-checklinkname=0` linker flag to suppress the warning

**Decision**: Upgrade sonic to v1.14.2

**Rationale**:
- The linker flag only suppresses the compile-time check, not the runtime warning
- Upgrading addresses the root cause rather than treating symptoms
- Sonic v1.14.2 is a minor version bump with backward compatibility guarantees
- The upgrade includes performance improvements and bug fixes
- No breaking changes to the sonic API

**Alternatives Considered**:
- Linker flag approach: Would still show runtime warning
- Downgrade Go version: Not feasible, Go 1.24+ is required
- Replace cosmos-sdk: Too invasive for this simple fix

**Outcome**: Successfully eliminated warning with zero side effects

---

## Decision 2: Verify Compatibility Without Modifying cosmos-sdk

**Date**: 2025-11-14

**Context**:
`cosmossdk.io/log@v1.5.0` declares a dependency on `sonic@v1.12.3`, but we're upgrading to v1.14.2. Need to verify this doesn't break cosmos-sdk.

**Decision**: Proceed with upgrade based on Go modules minimum version semantics

**Rationale**:
- Go modules use minimum version selection
- If a package requires v1.12.3, using v1.14.2 is safe (minor version compatibility)
- Sonic is not exposed in cosmossdk.io/log's public API
- Sonic is only used internally for JSON encoding in logs
- Semantic versioning guarantees v1.14.2 is backward compatible with v1.12.3

**Verification Performed**:
1. Dependency graph analysis: Confirmed sonic is 3 levels deep, indirect
2. API exposure check: Sonic not in cosmossdk.io/log public interface
3. Integration tests: All cosmos-sdk related tests passed
4. Functional tests: JSON serialization produces identical output
5. Build verification: Full project compiles without errors

**Outcome**: All verification passed; upgrade is fully compatible

---

## Decision 3: Minimal Change Approach

**Date**: 2025-11-14

**Context**:
Could have also updated cosmos-sdk or other related dependencies while making this change.

**Decision**: Only upgrade sonic, keep all other dependencies unchanged

**Rationale**:
- Follow single responsibility principle: one fix per PR
- Minimize risk surface area
- Easier to debug if issues arise
- Faster review process
- Clear git history showing exactly what changed

**Outcome**: Clean, focused change that's easy to understand and review
