# DefraDB - Claude Code Project Guide

## Project Overview

DefraDB is a peer-to-peer document database built in Go with CRDT-backed storage, GraphQL query interface, and multi-client support (Go, HTTP, CLI, C, JS/WASM).

## Code Comments

Write comments that explain **why**, not **what**. The code itself communicates what it does.

### Do NOT comment:
- Obvious operations (`// increment counter`, `// return the result`, `// loop through items`)
- Function calls that are self-descriptive (`// fetch the user`, `// save to database`)
- Variable assignments where the name already conveys intent
- Every step in a sequence — trust the reader to follow straightforward logic

### DO comment:
- **Why** a non-obvious approach was chosen over the simpler alternative
- Workarounds, hacks, or constraints imposed by external systems
- Business logic that isn't self-evident from code structure
- Performance-critical sections where the "obvious" approach was intentionally avoided
- Regex patterns, bitwise operations, or dense expressions that genuinely benefit from explanation
- TODO/FIXME with actionable context (not just `// TODO: fix this`)

### Docstrings:
- Write docstrings for public APIs, exported functions, and complex interfaces
- Skip docstrings on internal/private helpers where the signature is self-documenting
- Never restate the function name in prose (`getData - Gets the data`)
- Focus on: non-obvious parameters, edge cases, return value semantics, and side effects

### General rules:
- If a comment restates the next line of code in English, delete it
- Fewer high-value comments are better than many low-value comments
- A block of code needing heavy commenting is a signal to refactor, not to add more comments
- When editing existing code, do not add comments to unchanged lines

<!-- GSD:project-start source:PROJECT.md -->
## Project

**DefraDB Debug Skill**

A Claude Code skill (`/defradb:debug`) that agentically tests and debugs DefraDB through end-to-end black-box testing. Given a prompt describing an area of concern, the skill builds DefraDB (if stale), launches a fresh instance, then iteratively generates and executes GraphQL queries against the HTTP API to find and document bugs. It uses deep codebase understanding to guide query generation while reasoning from first principles about expected behavior.

**Core Value:** Find real bugs in DefraDB that unit and integration tests miss, by autonomously generating and executing targeted end-to-end workloads guided by both codebase understanding and independent correctness reasoning.

### Constraints

- **Interface**: HTTP GraphQL API only (no direct Go API calls) — true black-box testing
- **Instance**: Fresh instance per session, clean state — reproducibility
- **Store**: Memory by default — speed; other backends via flag
- **Build**: Go with CGO_ENABLED=1 required
- **Interrupts**: Only prompt user when anomaly is reproducible AND agent can articulate reasoning — minimize noise
- **Scope**: Single-node only for v1 — no P2P, no ACP
<!-- GSD:project-end -->

<!-- GSD:stack-start source:codebase/STACK.md -->
## Technology Stack

## Languages
- Go 1.25.5 - All application code (`go.mod` line 3)
- Rust/WASM - Schema migration lenses (`tests/lenses/`, built via `rustup target add wasm32-unknown-unknown`)
- JavaScript/WASM - Supports `GOOS=js GOARCH=wasm` build target for browser/JS environments (`Makefile` JS test targets)
## Runtime
- Go 1.25.5 (required by `go.mod`)
- CGO_ENABLED=1 (required, set in CI workflows)
- Go modules (`go.mod` / `go.sum`)
- Lockfile: `go.sum` present
## Frameworks
- `github.com/go-chi/chi/v5` v5.2.5 - HTTP routing (`http/router.go`)
- `github.com/spf13/cobra` v1.10.2 - CLI framework (`cli/cli.go`, `cmd/defradb/main.go`)
- `github.com/spf13/viper` v1.21.0 - Configuration management (`cli/`)
- `github.com/sourcenetwork/graphql-go` v0.7.4 (fork) - GraphQL schema/query engine (`internal/request/graphql/`)
- `github.com/wundergraph/graphql-go-tools/v2` v2.0.0-rc.246 - GraphQL tooling
- `github.com/stretchr/testify` v1.11.1 - Assertions and test suites
- `github.com/onsi/gomega` v1.39.1 - Matchers
- `github.com/sourcenetwork/testo` v0.2.0 - Custom test utilities
- `github.com/testcontainers/testcontainers-go` v0.40.0 - Container-based integration tests
- `gotest.tools/gotestsum` - Test runner (installed via `make deps:test`)
- `github.com/vektra/mockery/v3` v3.5.2 - Mock generation (config: `tools/configs/mockery.yaml`)
- `github.com/golangci/golangci-lint/v2` v2.3 - Linting (configs: `tools/configs/golangci.yaml`, `tools/configs/golangci-tests.yaml`)
- `yamllint` - YAML linting (config: `tools/configs/yamllint.yaml`)
- `golang.org/x/vuln/cmd/govulncheck` - Vulnerability scanning
- `golang.org/x/perf/cmd/benchstat` - Benchmark analysis
## Key Dependencies
- `github.com/sourcenetwork/corekv` v0.3.1 - Key-value storage abstraction layer (with sub-packages: `badger`, `blockstore`, `chunk`, `leveldb`, `memory`, `namespace`)
- `github.com/dgraph-io/badger/v4` v4.8.0 - Default persistent KV store backend (`node/store_badger.go`)
- `github.com/sourcenetwork/go-p2p` v0.1.9 - Peer-to-peer networking (`node/node_p2p.go`)
- `github.com/libp2p/go-libp2p` v0.47.0 - libp2p networking stack (underlying P2P)
- `github.com/sourcenetwork/acp_core` v0.8.1 - Access Control Policy core
- `github.com/sourcenetwork/sourcehub` v0.4.1 - SourceHub blockchain integration (ACP via Source Hub)
- `github.com/cosmos/cosmos-sdk` v0.53.5 - Cosmos SDK (used by SourceHub integration)
- `github.com/sourcenetwork/lens/host-go` v0.10.0 - Schema migration lens runtime (`internal/db/lens*.go`)
- `github.com/philippgille/chromem-go` v0.7.0 - Vector embedding/similarity search (`internal/db/embedding.go`)
- `github.com/sourcenetwork/goji` v0.0.9 - JSON handling utilities
- `github.com/ipfs/boxo` v0.37.0 - IPFS/IPLD content-addressed storage
- `github.com/ipfs/go-cid` v0.6.0 - Content identifiers
- `github.com/ipld/go-ipld-prime` v0.22.0 - IPLD data model
- `github.com/fxamacker/cbor/v2` v2.9.0 - CBOR encoding
- `github.com/valyala/fastjson` v1.6.4 - Fast JSON parsing
- `github.com/evanphx/json-patch/v5` v5.9.11 - JSON patch operations
- `go.opentelemetry.io/otel` v1.40.0 - OpenTelemetry tracing and metrics
- `go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp` v1.40.0 - OTLP metric export
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp` v1.40.0 - OTLP trace export
- `github.com/sourcenetwork/corelog` v0.0.8 - Structured logging
- `golang.org/x/crypto` v0.48.0 - Cryptographic primitives
- `github.com/lestrrat-go/jwx/v2` v2.1.6 - JWT/JWK handling
- `github.com/zalando/go-keyring` v0.2.6 - OS keyring integration
- `github.com/decred/dcrd/dcrec/secp256k1/v4` v4.4.1 - Elliptic curve cryptography
- `github.com/charmbracelet/bubbletea` v1.3.10 - TUI framework (wizard)
- `github.com/charmbracelet/bubbles` v1.0.0 - TUI components
- `github.com/charmbracelet/lipgloss` v1.1.0 - TUI styling
- `github.com/joho/godotenv` v1.5.1 - .env file loading
- `github.com/go-chi/cors` v1.2.2 - CORS middleware
- `github.com/getkin/kin-openapi` v0.133.0 - OpenAPI spec generation (`http/openapi.go`)
- `github.com/vito/go-sse` v1.1.3 - Server-Sent Events client (`http/client.go`)
- `google.golang.org/grpc` v1.78.0 - gRPC support
## Configuration
- `DEFRA_KEYRING_SECRET` - Keyring unlock password (required for keyring operations)
- `DEFRA_ENV=dev` - Development mode flag
- `DEFRA_CLIENT_HTTP`, `DEFRA_CLIENT_CLI`, `DEFRA_CLIENT_C`, `DEFRA_CLIENT_GO` - Test client type selection
- `DEFRA_MUTATION_TYPE` - Test mutation type (`gql`, `collection-named`)
- `DEFRA_DOCUMENT_ACP_TYPE` - ACP implementation (`source-hub`)
- `DEFRA_BADGER_MEMORY` - Use in-memory Badger for tests
- OpenAI API key (via environment, for embedding provider)
- `Makefile` - Primary build orchestration
- `tools/configs/golangci.yaml` - Go linter config
- `tools/configs/golangci-tests.yaml` - Test-specific linter config
- `tools/configs/mockery.yaml` - Mock generation config
- `tools/configs/yamllint.yaml` - YAML linter config
- `tools/configs/chglog/config.yml` - Changelog generation config
- `playground` - Include GraphQL playground (`playground/`)
- `telemetry` - Include OpenTelemetry instrumentation
- `change_detector` - Run change detection tests
- `npx` - Tests requiring npx
## Platform Requirements
- Go 1.25.5+
- CGO enabled (CGO_ENABLED=1)
- Rust toolchain (for lens WASM builds): `rustup target add wasm32-unknown-unknown`
- golangci-lint v2.3
- gotestsum
- mockery v3.5.2
- Debian Bookworm (slim) as Docker base image
- Ports: 9161, 9171, 9181
- Container image: `sourcenetwork/defradb` (Docker Hub and GHCR)
- Containerfile: `tools/defradb.containerfile`
- Build targets: Linux (amd64, arm, arm64), macOS, Windows
- Android NDK builds supported (`make build-c-shared-android`, min API level 21)
- C shared library builds (`make build-c-shared-linux`)
- WASM/JS builds (`GOOS=js GOARCH=wasm`)
## CI/CD
- `test-coverage.yml` - Test coverage on PRs to master/develop
- `lint.yml` - Linting
- `lint-then-benchmark.yml` - Lint + benchmarks
- `release.yml` - Multi-platform release (ubuntu, macOS, Windows)
- `docker-release.yml` - Docker image build/push (linux/amd64, linux/arm64)
- `check-tidy.yml` - go mod tidy verification
- `check-vulnerabilities.yml` - govulncheck
- `check-mocks.yml` - Mock freshness check
- `check-documentation.yml` - Doc verification
- `validate-containerfile.yml` - Container build validation
- `test-macos.yml` - macOS-specific tests
- `test-npx.yml` - NPX-tagged tests
- `test-limited-resource.yml` - Resource-constrained tests
- `build-then-deploy-ami.yml` - AMI deployment
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

## Naming Conventions
- Use `snake_case.go` for all Go source files
- Error definitions live in a dedicated `errors.go` per package (e.g. `internal/db/errors.go`, `client/errors.go`, `node/errors.go`)
- Test files use `_test.go` suffix per Go convention
- Mock files are auto-generated into `mocks/` subdirectories with snake_case filenames matching interface names
- Constructors use `New<Type>` pattern: `NewDB()`, `NewCollectionIndex()`, `newCollection()` (unexported for internal-only)
- Error constructors use `NewErr<Description>` pattern: `NewErrFailedToGetHeads()`, `NewErrInvalidStoredIndexKey()`
- Test functions use `Test<Feature><Scenario>` with descriptive camelCase: `TestGetCollectionByNameReturnsErrorGivenNonExistantCollection`
- Integration test helpers use `executeTestCase(t, test)` as a local wrapper per test package
- Error message strings use unexported `const` with `err` prefix: `errFailedToGetHeads string = "failed to get document heads"`
- Sentinel errors use exported `var` with `Err` prefix: `ErrLensCIDNotFound = errors.New(errLensCIDNotFound)`
- Package-level loggers: `var log = corelog.NewLogger("p2p")`
- Interfaces are defined in `client/` package (public API): `client.Collection`, `client.TxnStore`, `client.Txn`, `client.P2P`
- Implementation structs are unexported in `internal/`: `collection`, `DB`, `Txn`
- Interface compliance assertions: `var _ client.Collection = (*collection)(nil)` placed near struct definition
- Use short, lowercase names: `db`, `planner`, `fetcher`, `filter`, `mapper`
- Import aliases for disambiguation: `acpDB "github.com/sourcenetwork/defradb/internal/db/acp"`, `lensStore "github.com/sourcenetwork/lens/host-go/store"`
## Code Style
- `gofmt` with `-s` (simplify) flag enabled
- `goimports` for import organization
- Rewrite rules enforced: `interface{}` -> `any`, `a[b:len(a)]` -> `a[b:]`
- Max line length: 120 characters (enforced by `lll` linter)
- Config: `tools/configs/golangci.yaml`
- golangci-lint v2.3 with extensive config at `tools/configs/golangci.yaml`
- Separate test linter config at `tools/configs/golangci-tests.yaml` (enforces license headers on test files)
- YAML linting via yamllint with config at `tools/configs/yamllint.yaml`
- Enabled linters: `dupword`, `errcheck`, `errorlint`, `forbidigo`, `forcetypeassert`, `goconst`, `goheader`, `govet` (all analyzers except `shadow`, `fieldalignment`), `ineffassign`, `lll`, `nolintlint`, `revive`, `staticcheck`, `unconvert`, `unused`, `whitespace`
- `fmt.Print*` and `println` are **forbidden** via `forbidigo` -- use `corelog` instead
- `ioutil.*` is forbidden
- `nolint` directives must specify the linter being suppressed (`require-specific: true`)
- Run lint: `make lint`
- Fix lint: `make lint:fix`
- Order enforced by `goimports` with local prefixes:
- Dot imports are forbidden (`revive: dot-imports` rule)
- Duplicated imports are forbidden
- Source files: BSL license header (Business Source License) -- see `tools/configs/golangci.yaml` `goheader` section
- Test files: Dual AGPLv3/BSL license header -- see `tools/configs/golangci-tests.yaml`
- Format (source):
- Format (tests):
## Error Handling
- Use the project's `errors` package at `errors/errors.go`, NOT the standard `errors` package for creating new errors
- `errors.New(message, keyvals...)` -- creates error with stack trace
- `errors.Wrap(message, inner, keyvals...)` -- wraps an inner error with context
- `errors.NewKV(key, value)` -- creates key-value metadata for errors
- `errors.Is()`, `errors.As()` -- delegates to standard library
## Logging
- Create package-level logger: `var log = corelog.NewLogger("component-name")`
- Use structured logging with `corelog.Any(key, value)` for context
- Do NOT use `fmt.Print*` or `println` -- these are blocked by linter
## Interface & Constructor Patterns
- Always assert interface compliance at compile time near the struct definition:
- Exported constructors: `func New<Type>(params...) (*Type, error)` or `func New<Type>(params...) *Type`
- Unexported constructors for internal types: `func new<Type>(params...) (*type, error)`
- Factory functions where polymorphism needed: `fetcherFactory func() fetcher.Fetcher`
- Use builder pattern via `options` package: `options.Node().SetXxx()`
- `github.com/sourcenetwork/immutable` for optional values: `immutable.Option[T]`
## Documentation
- All exported functions/types should have doc comments starting with the name
- Error constructor functions document what the error indicates
- `@todo` and `@body` annotations for tracked technical debt items
- Reference GitHub issues in TODOs: `// TODO: https://github.com/sourcenetwork/defradb/issues/NNN`
- Conventional commits format: `<type>: <description>`
- Types: `feat`, `fix`, `tools`, `docs`, `refactor`, `test`, `ci`, `chore`, `bot`
- `<your-name>/<type>/<description>` (e.g. `jsimnz/feat/db-debug-skill`)
## Module Design
- Public API interfaces in `client/` package
- Internal implementations in `internal/` directory (Go internal package convention)
- Mocking via mockery with config at `tools/configs/mockery.yaml` -- generates into `mocks/` subdirectories
- Mocked interfaces: `Blockstore`, `Txn`, `Collection`, `Fetcher`, `EncodedDocument`, `DB`, `P2P`, `CommChannel`
- Environment variables for runtime config (prefixed with `DEFRA_`)
- Build tags for conditional compilation (`change_detector`, `npx`, `playground`)
- Go version: 1.25 (specified in `tools/configs/golangci.yaml`)
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

## Pattern Overview
- Single binary monolith with optional P2P and HTTP API subsystems
- CRDT-based conflict-free replication using Merkle DAG structures
- GraphQL as the primary query language (parsed internally, not via standard GQL server)
- Transaction-based storage with multi-store namespace separation
- Access Control Policy (ACP) system at both document and node levels
- Schema migration via Lens transforms (WASM-based)
## Layers
- Purpose: Provides user-facing entry points (CLI commands, HTTP API)
- Location: `cli/`, `http/`
- Contains: Cobra CLI commands (`cli/start.go`, `cli/request.go`, etc.), HTTP handlers and router (`http/handler.go`, `http/router.go`)
- Depends on: `client` interfaces, `node` package
- Used by: End users and external HTTP clients
- Purpose: Orchestrates subsystem lifecycle (DB, P2P, HTTP API)
- Location: `node/node.go`
- Contains: Node struct that composes DB, Peer, and Server; startup/shutdown logic; store initialization
- Depends on: `internal/db`, `http`, P2P subsystem, ACP providers
- Used by: `cli/start.go`
- Purpose: Defines the public API contract for the database
- Location: `client/`
- Contains: `TxnStore` interface (`client/db.go`), `Store` interface, `Collection` interface (`client/collection.go`), `Document` type (`client/document.go`), request types (`client/request/`)
- Depends on: Nothing internal (pure interface + types)
- Used by: All other layers; this is the primary boundary
- Purpose: Core database logic - schema management, document CRUD, query execution, merge/sync
- Location: `internal/db/`
- Contains: `DB` struct (`internal/db/db.go`), collection implementation (`internal/db/collection.go`), document operations (`internal/db/document.go`, `internal/db/document_get.go`, `internal/db/document_update.go`, `internal/db/document_delete.go`), request execution (`internal/db/request.go`), merge logic (`internal/db/merge.go`)
- Depends on: `internal/planner`, `internal/datastore`, `internal/core`, `internal/request/graphql`, `internal/db/fetcher`
- Used by: `node`, `http` (via `client` interfaces)
- Purpose: Translates parsed GraphQL requests into executable plan trees
- Location: `internal/planner/`
- Contains: `Planner` struct (`internal/planner/planner.go`), plan nodes (`selectNode`, `scanNode`, `updateNode`, `deleteNode`, `groupNode`, `orderNode`, `limitNode`, `typeIndexJoin`, etc.), mapper (`internal/planner/mapper/`), filter logic (`internal/planner/filter/`)
- Depends on: `client`, `internal/core`, `internal/db/fetcher`, `internal/keys`
- Used by: `internal/db/request.go`
- Purpose: Parses GraphQL SDL schemas and query strings into typed ASTs
- Location: `internal/request/graphql/`
- Contains: GraphQL parser implementation, schema management
- Depends on: `client/request`, `internal/core` (implements `core.Parser` interface)
- Used by: `internal/db` (via `core.Parser`)
- Purpose: Provides transactional, namespaced key-value storage abstraction
- Location: `internal/datastore/`
- Contains: `Multistore` (`internal/datastore/multi.go`) with namespaced sub-stores (system, data, head, block, peer, enc), `Txn` wrapper (`internal/datastore/txn.go`), blockstore (`internal/datastore/blockstore.go`)
- Depends on: `corekv` (external KV store interface), `internal/db/lock`
- Used by: `internal/db`, `internal/planner`
- Purpose: Shared internal types, CRDTs, block structures
- Location: `internal/core/`
- Contains: `Doc` type (`internal/core/doc.go`), `Parser` interface (`internal/core/parser.go`), CRDT types (`internal/core/crdt/`), block types (`internal/core/block/`), CID utilities (`internal/core/cid/`)
- Depends on: `client`
- Used by: All internal packages
- Purpose: Defines storage key types for all namespaced stores
- Location: `internal/keys/`
- Contains: Key types for datastore, headstore, peerstore, systemstore (e.g., `internal/keys/datastore_doc.go`, `internal/keys/headstore.go`, `internal/keys/systemstore_collection.go`)
- Depends on: Nothing significant
- Used by: `internal/db`, `internal/planner`, `internal/datastore`
- Purpose: Document-level (DAC) and node-level (NAC) access control
- Location: `acp/` (public interfaces), `internal/db/acp/` (internal implementation)
- Contains: DAC interface (`acp/dac/`), NAC interface (`acp/nac/`), local ACP provider (`acp/local/`), identity types (`acp/identity/`)
- Depends on: External SourceHub for non-local ACP
- Used by: `internal/db`, `internal/planner`, `node`
## Data Flow
- All state is stored in a single `corekv.TxnStore` (rootstore), namespaced into sub-stores via `Multistore`
- Transactions wrap the rootstore and provide isolation via `internal/datastore.Txn`
- Context-based transaction propagation: `ensureContextTxn()` in `internal/db/context.go`
- Event bus (`event.Bus` in `event/`) provides async notification of mutations
## Key Abstractions
- Purpose: Represents a node in the query execution plan tree
- Examples: `internal/planner/scan.go` (scanNode), `internal/planner/select.go` (selectNode), `internal/planner/type_join.go` (typeIndexJoin)
- Pattern: Iterator pattern with `Init()`, `Start()`, `Next()`, `Value()`, `Close()`
- Purpose: Primary public API for interacting with DefraDB
- Examples: `client/db.go`
- Pattern: Interface-based; `internal/db.DB` is the concrete implementation
- Purpose: Provides logical separation of storage concerns on a single KV store
- Examples: `internal/datastore/multi.go`
- Pattern: Namespace-prefixed sub-stores: `s` (system), `d` (data), `h` (head), `b` (block), `p` (peer), `e` (encrypted)
- Purpose: Decouples query language parsing from the database engine
- Examples: `internal/core/parser.go` (interface), `internal/request/graphql/parser.go` (implementation)
- Pattern: Strategy pattern allowing different query language backends
- Purpose: Conflict-free replicated data types for each field type
- Examples: `internal/core/crdt/lww.go` (Last-Writer-Wins register), `internal/core/crdt/composite.go`, `internal/core/crdt/counter.go`
- Pattern: Each document field maps to a CRDT type; mutations produce deltas stored as IPLD blocks
## Entry Points
- Location: `cmd/defradb/main.go`
- Triggers: User runs `defradb` command
- Responsibilities: Creates Cobra root command via `cli.NewDefraCommand()`, delegates to subcommands
- Location: `cli/start.go` -> `node/node.go` `Start()`
- Triggers: `defradb start` CLI command
- Responsibilities: Creates rootstore, initializes ACP, starts P2P, creates DB instance, starts HTTP API
- Location: `http/handler.go`, `http/server.go`
- Triggers: HTTP requests to port 9181 (default)
- Responsibilities: Routes requests to appropriate handlers (store, collection, document, tx, p2p, acp, block)
- Location: `internal/db/db.go` `NewDB()`
- Triggers: Go code importing the package
- Responsibilities: Creates a DB instance directly without HTTP/P2P (used in tests and embedded scenarios)
## Error Handling
- Errors are defined per-package in `errors.go` files using constructor functions (e.g., `NewErrDocNotFound()`)
- Error wrapping preserves context: `errors.Wrap("failed to do X", err)`
- HTTP layer maps errors to appropriate HTTP status codes
- Transaction errors trigger retry logic (configurable `MaxTxnRetries`)
## Cross-Cutting Concerns
<!-- GSD:architecture-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd:quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd:debug` for investigation and bug fixing
- `/gsd:execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->

<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd:profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
