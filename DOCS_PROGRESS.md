# Documentation Audit Progress

## Overview
Systematic comparison of `docs/website/` content against actual codebase behavior.
**All research and fix phases COMPLETE.** 22 files modified, 70+ inconsistencies found and fixed.

## Summary

| Area | Files | Issues Found | Status |
|------|-------|-------------|--------|
| Getting Started | 1 | 14 | FIXED |
| Query Specification | 8 | 41 | FIXED |
| Guides | 7 | 16 | FIXED |
| Concepts | 1 | 3 | FIXED |
| CLI References | 1 (+1 source) | 1 | FIXED |
| HTTP/OpenAPI | 2 source files | 3 fixed, 8 noted | PARTIAL |

**Total: 22 files changed, ~230 insertions, ~187 deletions**

## Work Areas

### 1. Getting Started (`getting-started.md`) — FIXED
- `defradb client ping` → `curl localhost:9181/api/v0/health`
- `_ge` filter → `_geq`
- Mutation response JSON format corrected (operation name wrapper)
- `_commits` links field `name` → `fieldName`
- `--tcpaddr` flag removed (gRPC server no longer exists)
- `defradb client rpc` → `client p2p` (all occurrences)
- p2p collection add: CIDs → collection names
- Replicator command syntax fully updated
- `--tls` flag removed; TLS via `--pubkeypath`/`--privkeypath` only
- TLS key descriptions corrected (server.key=private, server.crt=cert)
- Let's Encrypt/autocert section removed (not implemented)
- `privkeypaths` typo fixed
- Peer addresses `0.0.0.0` → `127.0.0.1`

### 2. Query Specification — ALL 8 FILES FIXED

#### filtering.md
- `_gte` → `_geq`, `_lte` → `_leq` throughout
- Table 2 String operators: added `_nlike`, `_ilike`, `_nilike`
- Implicit scalar shorthand note corrected
- New Array Operators section (`_any`, `_all`, `_none`)
- `Integer` → `Int`, `Floating Point` → `Float`
- `_ilike`/`_nilike` added to Table 1

#### mutation-block.md
- `dockey` → `docID` everywhere
- `ids: [ID]` removed; `docID: [ID!]` used instead
- Single quotes → double quotes, `nil` → `null`
- Input type names corrected (`BookMutationInputArg`)
- `add` mutation input: single → list
- Hard delete → soft delete (corrected)
- `_le` → `_leq`
- Upsert section added

#### limiting-and-pagination.md
- `sort:` → `order:`

#### aliases.md
- `sort:` → `order:`

#### sorting-and-ordering.md
- `_id` filter removed; `docID` arg used
- Integer IDs → string `bae-...`
- `DocKey` → `_docID`

#### aggregate-functions.md
- Syntax `AVG { rating }` → `AVG(GROUP: {field: rating})`
- "Table 3" references removed
- Underscore prefix removed from function names

#### execution-flow.md
- `having` clause removed (not implemented)

#### database-api.md
- Commit type fully rewritten (correct fields: delta as String, links/heads as [Commit], signature, docID, etc.)
- `CommitLink` and `Delta` types removed
- `dockey` → `docID`, single quotes → double quotes
- `cid` argument corrected to list `[ID!]`

#### relationships.md
- `@embed` directive removed (doesn't exist)
- `Integer` → `Int`
- `@relation` positional args → `name:` kwarg
- `DocKey` → document ID

### 3. Guides — ALL 7 FILES FIXED

#### schema-relationship.md
- 7 capitalized type names → lowercase field names in queries
- `update_Author(id:` → `update_Author(docID:`

#### explain-systems.md
- Simple Explain: `"select TopNode"` → `"operationNode"` array
- `collectionID` numeric → CID hash
- `"spans"` → `"prefixes"`
- Execute Explain: `selectTopNode` inside `operationNode` array
- `filterMatches` moved to selectNode (not scanNode)

#### peer-to-peer.md
- `client rpc addreplicator` → `client p2p replicator add -c`
- Default P2P addr `0.0.0.0` → `127.0.0.1`
- Replicator persistence claim corrected

#### collection-migration.md
- `--set-active=true` → separate `collection set-active` command
- Code fence labels `graphql` → `bash`/`rust`

#### time-traveling-queries.md
- `dockey` → `docID`

#### deployment-guide.md
- Go version "up to 1.20" → requires 1.25.5
- P2P disable claim corrected

#### akash-deployment.md
- `p2p replicator set` → `p2p replicator add`
- JSON body → positional multiaddr + `-c` flag
- `p2p info` response format corrected

### 4. Concepts — FIXED

#### ipfs.md
- "directed acrylic graphs" → "directed acyclic graphs"
- Added "How DefraDB Uses IPFS Technologies" section
- Clarified DefraDB uses IPFS libraries, not full IPFS network

### 5. CLI References — FIXED
- `cli/acp_nac.go` + `defradb_client_acp_node.md`: `--acp-enable true` → `--node-acp-enable`

### 6. HTTP/OpenAPI — PARTIALLY FIXED

**Fixed in source code:**
- `handler_p2p.go`: OpenAPI schema `collectionName` → `collectionID`
- `handler_acp.go`: 8 operationIDs spaces → snake_case

**Noted (not fixed — require more extensive changes):**
- Missing `requestBody` for `PATCH /collections/{name}/document/{docID}`
- Missing `operationName`/`variables` query params for `GET /graphql`
- Missing `show_deleted` query param for `GET /collections/{name}/document/{docID}`
- Missing `encrypt`/`encryptFields` query params for `POST /collections/{name}`
- 12 tags used but not defined; `acp` tag defined but unused
- `bearerToken` security scheme defined but never applied
- `txn` common parameter defined but never referenced
- 3 P2P endpoints missing descriptions

## Not Analyzed (Low Priority)
- `docs/website/references/query-specification/query-block.md` — needs review
- `docs/website/references/query-specification/collections.md` — needs review
- `docs/website/concepts/libp2p.md` — needs review (similar issues to ipfs.md likely)
- `docs/website/guides/merkle-crdt.md` — needs review
- `docs/website/references/query-specification/query-language-overview.md` — needs review
- Release notes — historical, not audited
