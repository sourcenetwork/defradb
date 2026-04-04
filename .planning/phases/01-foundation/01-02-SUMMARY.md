---
phase: 01-foundation
plan: 02
subsystem: testing
tags: [graphql, http-api, integration-test, defradb]

requires:
  - phase: 01-01
    provides: SKILL.md with lifecycle and query instructions
provides:
  - Verified end-to-end SKILL.md commands against real DefraDB instance
  - Battle-tested corrections (mutation prefix, portable paths, session isolation, batch fixtures)
affects: [02-correctness-engine]

tech-stack:
  added: []
  patterns: [black-box HTTP testing via curl + jq]

key-files:
  created: []
  modified:
    - .claude/skills/defradb-debug/SKILL.md

key-decisions:
  - "GraphQL mutation prefix is add_ not create_ — confirmed by DefraDB error message"
  - "Use git rev-parse --show-toplevel for portable project root detection"
  - "Session files keyed by port (/tmp/.defradb-debug-session-${PORT}) to support concurrent sessions"
  - "Batch fixture loading via array input to add_ mutations instead of one-at-a-time"

patterns-established:
  - "Validate skill commands against real instance before shipping"

requirements-completed: [LIFE-02, LIFE-03, LIFE-04, QEXE-01]

duration: 12min
completed: 2026-04-04
---

# Plan 01-02: Integration Test Summary

**Full CRUD cycle validated against live DefraDB — fixed mutation prefix, hardcoded paths, session collisions, and one-at-a-time fixture loading**

## Performance

- **Duration:** 12 min
- **Started:** 2026-04-04
- **Completed:** 2026-04-04
- **Tasks:** 2 (1 automated + 1 human verification)
- **Files modified:** 1

## Accomplishments
- Ran complete build -> start -> health check -> schema -> mutation -> query -> shutdown cycle
- Discovered and fixed `create_` -> `add_` mutation prefix mismatch
- Fixed hardcoded project path to use `git rev-parse --show-toplevel`
- Changed session files from global singleton to per-port isolation
- Converted one-at-a-time fixture loading to batch array input

## Task Commits

1. **Task 1: Build and start DefraDB, execute full CRUD cycle** - `286a91fb5` (fix)
2. **Task 2: Human verification + 3 additional fixes** - `d4a242955` (fix)

## Files Created/Modified
- `.claude/skills/defradb-debug/SKILL.md` - Fixed mutation prefix, portable paths, session isolation, batch fixtures

## Decisions Made
- Mutation prefix is `add_` not `create_` — DefraDB's error message suggested the correct alternatives
- Port-keyed session files prevent concurrent debug session collisions
- `add_` mutations accept `input: [...]` arrays (confirmed in schema generator at `internal/request/graphql/schema/generate.go:1224`)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Wrong GraphQL mutation prefix**
- **Found during:** Task 1 (CRUD cycle)
- **Issue:** SKILL.md used `create_<Collection>` but DefraDB expects `add_<Collection>`
- **Fix:** Updated fixture loading and documentation
- **Files modified:** .claude/skills/defradb-debug/SKILL.md
- **Verification:** `add_User` mutation returned valid `_docID`
- **Committed in:** 286a91fb5

**2. [User feedback] Three portability and efficiency issues**
- **Found during:** Task 2 (human review)
- **Issue:** Hardcoded path, global session file, one-at-a-time fixture inserts
- **Fix:** Portable root detection, per-port session files, batch array mutations
- **Files modified:** .claude/skills/defradb-debug/SKILL.md
- **Verification:** Code review confirmed all references updated
- **Committed in:** d4a242955

---

**Total deviations:** 2 (1 auto-fixed bug, 1 user-directed improvement)
**Impact on plan:** All fixes improve correctness and portability. No scope creep.

## Issues Encountered
None beyond the deviations above.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- SKILL.md is complete and validated — ready for Phase 2 correctness engine
- All lifecycle commands produce correct results against real DefraDB
- Fixture loading is efficient (batch) and uses correct mutation API

---
*Phase: 01-foundation*
*Completed: 2026-04-04*
