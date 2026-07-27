<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/DefraDB_White.svg">
    <img height="80px" alt="DefraDB" src="docs/DefraDB_Full.svg">
  </picture>
</p>

<h3 align="center">🏠 Internal Contributor Guide</h3>

<p align="center">
  <sub>This document supplements the main <a href="./CONTRIBUTING.md">CONTRIBUTING.md</a> with internal processes, CI details, and project conventions.</sub>
  <br>
  <sub><strong>📖 Read the main contributing guide first</strong> - everything here builds on top of it.</sub>
</p>

---

<details>
<summary><strong>📑 Table of Contents</strong> <em>(click to expand)</em></summary>

&nbsp;

- [🔀 Branch Flow](#-branch-flow)
- [💬 Commenting Etiquette](#-commenting-etiquette)
- [⚙️ CI Checks - Complete Reference](#-ci-checks--complete-reference)
  - [✅ Required CI Checks](#-required-ci-checks)
  - [📊 Additional CI Checks (Non-Blocking)](#-additional-ci-checks-non-blocking)
  - [🔧 Quick Fix Checklist](#-quick-fix-checklist)
- [🧪 Testing - Advanced](#-testing--advanced)
  - [🏃 Test Configuration Variables](#-test-configuration-variables)
  - [🌐 SourceHub ACP Tests](#-sourcehub-acp-tests)
  - [📈 Benchmarks](#-benchmarks)
  - [🔍 Change Detector](#-change-detector)
- [📦 Dependency Management](#-dependency-management)
  - [🐹 Go Version Bumping Policy](#-go-version-bumping-policy)
- [💡 Ideation / Proposals (SIPs)](#-ideation--proposals-sips)
- [📋 Project Management](#-project-management)
- [⚖️ Licensing](#-licensing)
- [⚠️ Breaking Changes](#-breaking-changes)

</details>

---

## 🔀 Branch Flow

> [!NOTE]
> Fork Flow is **always preferred**, even for internal developers. Branch Flow should only be used when direct repository access is needed (e.g., CI workflow updates).

In specific cases, internal developers may use the Branch Flow instead of Fork Flow:

| Step | Action |
|------|--------|
| 1️⃣ | Clone the main repository locally |
| 2️⃣ | Create a feature branch following the naming convention below |
| 3️⃣ | Make changes on the feature branch |
| 4️⃣ | Run all required local checks |
| 5️⃣ | Open a pull request targeting `develop` |

**Branch naming convention:**

```text
<dev-name>/<label>/<description>
```

<details>
<summary>📋 <strong>Examples</strong></summary>

&nbsp;

| Branch Name | Purpose |
|-------------|---------|
| `lone/ci/update-test-workflow-action` | CI workflow update |
| `lone/feat/add-encryption-support` | New feature |
| `lone/fix/resolve-query-timeout` | Bug fix |
| `andrew/refactor/simplify-p2p-sync` | Refactoring |

</details>

**Run all checks before opening your PR:**

```shell
make docs && make mocks && make lint && make tidy && make test && make test:changes
```

---

## 💬 Commenting Etiquette

We use labels inspired by **[Conventional Comments](https://conventionalcomments.org/)** to clarify the nature and urgency of each review comment. Prefix your review comments with the label.

| &nbsp; | Label | Meaning | Action Required |
|--------|-------|---------|----------------|
| 💭 | `thought` | A dump of thoughts - may or may not be within scope. Provides context or sparks ideas. | No action required. |
| ❓ | `question` | A question from the reviewer. May evolve into other types once clarity is achieved. | Answer the question, or point to a helpful resource. |
| 🔍 | `nitpick` | Minor, nitpicky suggestion. | Can be ignored or accepted - no follow-up required. |
| 💡 | `suggestion` | Non-blocking suggestion. | Accept it, or explain why it shouldn't be done. |
| 📋 | `todo` | **Blocking** - must be resolved before merge. | Must resolve before merge. If deferring, create an issue and link it. |

<details>
<summary>📋 <strong>Example review comments</strong></summary>

&nbsp;

```text
suggestion: Consider using a map here instead of a slice for O(1) lookups.
```

```text
todo: This needs error handling - if the connection drops mid-sync we'll panic.
```

```text
thought: We might want to consider extracting this into its own package
if we end up reusing it across the codebase.
```

</details>

---

## ⚙️ CI Checks - Complete Reference

When a PR is opened, our CI pipeline runs a comprehensive suite of checks. Here's the full breakdown.

### ✅ Required CI Checks

These run on every PR targeting `develop` or `master` and **must pass** to merge:

| Check | What It Does | Fix Locally |
|-------|--------------|-------------|
| 🔗 **Build C Shared Library (Linux)**<br>[`build-c-shared-linux.yml`](./.github/workflows/build-c-shared-linux.yml) | Builds and tests the Linux C shared library (`libdefradb`) | `make build-c-shared-linux` |
| 🏗️ **Build Dependencies**<br>[`build-dependencies.yml`](./.github/workflows/build-dependencies.yml) | Ensures all project dependencies can be built | `make deps` |
| 📊 **Check Data Format Changes**<br>[`check-data-format-changes.yml`](./.github/workflows/check-data-format-changes.yml) | Detects backwards-incompatible data format changes. Must be documented in [`docs/data_format_changes/`](./docs/data_format_changes/README.md) | `make test:changes` *(see [change detector README](./tests/change_detector/README.md))* |
| 📖 **Check Documentation**<br>[`check-documentation.yml`](./.github/workflows/check-documentation.yml) | Ensures CLI docs, HTTP API docs, and README TOC are up to date *(3 sub-checks: cli, http, readme-toc)* | `make docs` *(runs `docs:cli`, `docs:http`, `toc`)* |
| 🔧 **Check Mocks**<br>[`check-mocks.yml`](./.github/workflows/check-mocks.yml) | Verifies all mocks are regenerated and up to date | `make mocks` |
| 📦 **Check Tidy**<br>[`check-tidy.yml`](./.github/workflows/check-tidy.yml) | Ensures `go.mod` and `go.sum` are clean | `make tidy` |
| 🔒 **Check Vulnerabilities**<br>[`check-vulnerabilities.yml`](./.github/workflows/check-vulnerabilities.yml) | Runs `govulncheck` to scan for known security vulnerabilities | `make deps:vulncheck && govulncheck ./...` |
| 🧙 **Check Wizard Health**<br>[`check-wizard-health.yml`](./.github/workflows/check-wizard-health.yml) | Tests the interactive setup wizard using an automated expect script | `make build && ./tools/scripts/wizard_test.sh` |
| 🧹 **Lint**<br>[`lint.yml`](./.github/workflows/lint.yml) | Runs **golangci-lint** ([config](./tools/configs/golangci.yaml)) and **yamllint** ([config](./tools/configs/yamllint.yaml)) | `make deps:lint && make lint` *(auto-fix: `make lint:fix`)* |
| 🧹📊 **Lint Then Benchmark**<br>[`lint-then-benchmark.yml`](./.github/workflows/lint-then-benchmark.yml) | Linting + conditional benchmarks. *`SHORT` for PRs to develop · `FULL` with label · Skip with `action/no-benchmark`* | `make lint` then `make test:bench-short` |
| 🚀 **Start Binary**<br>[`start-binary.yml`](./.github/workflows/start-binary.yml) | Builds the binary and verifies it starts | `make build && ./build/defradb start --no-keyring` |
| 🧪 **Test Coverage**<br>[`test-coverage.yml`](./.github/workflows/test-coverage.yml) | Comprehensive test matrix: clients (Go/HTTP/CLI), databases, mutations, ACP, lenses, views, encryption, vectors → [Codecov](https://codecov.io/gh/sourcenetwork/defradb) | `make test` with [env variables](#-test-configuration-variables) |
| 🐧 **Test Debian Package**<br>[`test-deb-package.yml`](./.github/workflows/test-deb-package.yml) | Builds the `libdefradb` Debian package, installs it, and verifies it works | `make build-c-shared-linux:deb` |
| 🐋 **Validate Containerfile**<br>[`validate-containerfile.yml`](./.github/workflows/validate-containerfile.yml) | Builds Docker image from [containerfile](./tools/defradb.containerfile) and verifies it runs | Ensure [containerfile](./tools/defradb.containerfile) is valid |
| 🏷️ **Validate Title**<br>[`validate-title.yml`](./.github/workflows/validate-title.yml) | Validates PR title follows our [title format rules](./CONTRIBUTING.md#-title-format) (inspired by conventional commits) | Fix title per [rules](./CONTRIBUTING.md#-title-format) · [script](./tools/scripts/validate-conventional-style.sh) |

### 📊 Additional CI Checks (Non-Blocking)

These also run on PRs but are **informational** - failures won't block merge:

| Check | What It Does |
|-------|--------------|
| 🐢 **Test Limited Resource**<br>[`test-limited-resource.yml`](./.github/workflows/test-limited-resource.yml) | Runs tests on standard (slower) runners to catch resource-constrained failures |
| 🍎 **Test macOS**<br>[`test-macos.yml`](./.github/workflows/test-macos.yml) | Integration tests on macOS for cross-platform compatibility |
| 📜 **Test NPX/JS Build**<br>[`test-npx.yml`](./.github/workflows/test-npx.yml) | Verifies NPX/JavaScript-dependent tests can build and run |
| ☁️ **Preview AMI**<br>[`preview-ami-with-terraform-plan.yml`](./.github/workflows/preview-ami-with-terraform-plan.yml) | Triggers on AWS infra changes only - validates Terraform and posts plan as PR comment |

### 🔧 Quick Fix Checklist

> [!TIP]
> If CI is failing, try these locally:

```shell
make tidy          # Fix go.mod/go.sum issues
make docs          # Regenerate documentation
make mocks         # Regenerate mocks
make lint          # Check for lint errors
make lint:fix      # Auto-fix lint errors where possible
make test          # Run the full test suite
make test:changes  # Check for data format changes
```

---

## 🧪 Testing - Advanced

### 🏃 Test Configuration Variables

The test suite uses environment variables to control which configurations are tested:

| Variable | Values | Description |
|----------|--------|-------------|
| `DEFRA_CLIENT_GO` | `true`/`false` | Enable Go client tests |
| `DEFRA_CLIENT_HTTP` | `true`/`false` | Enable HTTP client tests |
| `DEFRA_CLIENT_CLI` | `true`/`false` | Enable CLI client tests |
| `DEFRA_BADGER_MEMORY` | `true`/`false` | Use in-memory Badger store |
| `DEFRA_BADGER_FILE` | `true`/`false` | Use file-based Badger store |
| `DEFRA_BADGER_ENCRYPTION` | `true`/`false` | Enable Badger encryption |
| `DEFRA_MUTATION_TYPE` | `gql` / `collection-named` / `collection-save` | Mutation type |
| `DEFRA_DOCUMENT_ACP_TYPE` | `local` / `source-hub` | ACP type |
| `DEFRA_LENS_TYPE` | `wasm-time` / `wasm-er` | Lens WASM runtime |
| `DEFRA_VIEW_TYPE` | `cacheless` / `materialized` | View type |
| `DEFRA_VECTOR_EMBEDDING` | `true`/`false` | Enable vector embedding tests |

### 🌐 SourceHub ACP Tests

> [!WARNING]
> SourceHub ACP tests require **Docker** and are resource-heavy.

```shell
DEFRA_CLIENT_HTTP=true DEFRA_CLIENT_GO=false DEFRA_DOCUMENT_ACP_TYPE=source-hub \
  DEFRA_BADGER_MEMORY=true go test ./tests/integration/acp/... -count=1 -timeout 20m
```

Use `-p 1` when running the full suite to avoid Docker resource contention.

### 📈 Benchmarks

**Running benchmarks:**

```shell
make test:bench          # Full benchmark suite
make test:bench-short    # Short benchmark suite
```

**Comparing against `develop`:**

```shell
# On develop branch
make test:bench | tee develop.txt

# On your feature branch
make test:bench | tee current.txt

# Compare results
make deps:bench     # Install benchstat
benchstat develop.txt current.txt
```

**CI benchmark labels:**

| Label | Effect |
|-------|--------|
| `action/full-benchmark` | Triggers a full benchmark run on the PR |
| `action/no-benchmark` | Skips benchmarks entirely |

### 🔍 Change Detector

The data format change detector (`make test:changes`) ensures backwards compatibility. If data format changes are detected, they must be documented in [`docs/data_format_changes/`](./docs/data_format_changes/README.md).

> 📖 See the [change detector README](./tests/change_detector/README.md) for details on how it works.

---

### 🐹 Go Version Bumping Policy

| &nbsp; | Policy | Details |
|--------|--------|---------|
| 🎯 | **One version behind** | We use one version behind the latest Go release. A Go release becomes unsupported when the second new major version is released after it. |
| 🔒 | **Security exceptions** | If `govulncheck` reports vulnerabilities fixed only in the latest Go version (and patches haven't landed on our current version within ~24 hours), we do **not** bump preemptively. See the [vulnerability check workflow](./.github/workflows/check-vulnerabilities.yml). |
| 📦 | **Dependency-driven bumps** | If a dependency strictly requires a newer Go version and DefraDB can't resolve it otherwise, we bump accordingly. |

---

## 💡 Ideation / Proposals (SIPs)

For significant architectural changes or major new features, write a **[Source Improvement Proposal (SIP)](https://github.com/sourcenetwork/SIPs/)** to get community and team feedback before implementation.

---

## 📋 Project Management

- Use [Milestones](https://github.com/sourcenetwork/defradb/milestones) and the [project board](https://github.com/orgs/sourcenetwork/projects/3) to coordinate work on releases.

---

## ⚖️ Licensing

- Include the [BSL license header](./licenses/BSL.txt) at the top of **every** new code file.
- DefraDB is released under the [Business Source License (BSL)](./licenses/BSL.txt). Each dated version converts to Apache License v2.0 after four years.

---

## ⚠️ Breaking Changes

> [!IMPORTANT]
> When introducing breaking changes:
>
> 1. Include the `BREAKING CHANGE` keyword in the **commit message body** (not the title)
> 2. Follow it with a description of what changed and why
> 3. Document the changes in [`docs/data_format_changes/`](./docs/data_format_changes/) for the change detector to pass

---

<p align="center">
  <sub>📖 For the main contribution guide, see <a href="./CONTRIBUTING.md">CONTRIBUTING.md</a></sub>
</p>
