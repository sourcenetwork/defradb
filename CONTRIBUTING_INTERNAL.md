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
- [💬 Commenting Etiquette - Detailed Guide](#-commenting-etiquette--detailed-guide)
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

```
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

## 💬 Commenting Etiquette - Detailed Guide

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

```
suggestion: Consider using a map here instead of a slice for O(1) lookups.
```

```
todo: This needs error handling - if the connection drops mid-sync we'll panic.
```

```
thought: We might want to consider extracting this into its own package
if we end up reusing it across the codebase.
```

</details>

---

## ⚙️ CI Checks - Complete Reference

When a PR is opened, our CI pipeline runs a comprehensive suite of checks. Here's the full breakdown.

### ✅ Required CI Checks

These run on every PR targeting `develop` or `master` and **must pass** to merge:

<table>
<tr><th>Check</th><th>What It Does</th><th>Fix Locally</th></tr>

<tr>
<td>🏗️ <strong>Build Dependencies</strong><br><sub><a href="./.github/workflows/build-dependencies.yml">build-dependencies.yml</a></sub></td>
<td>Ensures all project dependencies can be built</td>
<td><code>make deps</code></td>
</tr>

<tr>
<td>📊 <strong>Check Data Format Changes</strong><br><sub><a href="./.github/workflows/check-data-format-changes.yml">check-data-format-changes.yml</a></sub></td>
<td>Detects backwards-incompatible data format changes. Must be documented in <a href="./docs/data_format_changes/README.md"><code>docs/data_format_changes/</code></a></td>
<td><code>make test:changes</code><br><sub>See <a href="./tests/change_detector/README.md">change detector README</a></sub></td>
</tr>

<tr>
<td>📖 <strong>Check Documentation</strong><br><sub><a href="./.github/workflows/check-documentation.yml">check-documentation.yml</a></sub></td>
<td>Ensures CLI docs, HTTP API docs, and README TOC are up to date<br><sub>(3 sub-checks: cli, http, readme-toc)</sub></td>
<td><code>make docs</code><br><sub>(runs <code>docs:cli</code>, <code>docs:http</code>, <code>toc</code>)</sub></td>
</tr>

<tr>
<td>🔧 <strong>Check Mocks</strong><br><sub><a href="./.github/workflows/check-mocks.yml">check-mocks.yml</a></sub></td>
<td>Verifies all mocks are regenerated and up to date</td>
<td><code>make mocks</code></td>
</tr>

<tr>
<td>📦 <strong>Check Tidy</strong><br><sub><a href="./.github/workflows/check-tidy.yml">check-tidy.yml</a></sub></td>
<td>Ensures <code>go.mod</code> and <code>go.sum</code> are clean</td>
<td><code>make tidy</code></td>
</tr>

<tr>
<td>🔒 <strong>Check Vulnerabilities</strong><br><sub><a href="./.github/workflows/check-vulnerabilities.yml">check-vulnerabilities.yml</a></sub></td>
<td>Runs <code>govulncheck</code> to scan for known security vulnerabilities</td>
<td><code>make deps:vulncheck && govulncheck ./...</code></td>
</tr>

<tr>
<td>🧙 <strong>Check Wizard Health</strong><br><sub><a href="./.github/workflows/check-wizard-health.yml">check-wizard-health.yml</a></sub></td>
<td>Tests the interactive setup wizard using an automated expect script</td>
<td><code>make build && ./tools/scripts/wizard_test.sh</code></td>
</tr>

<tr>
<td>🧹 <strong>Lint</strong><br><sub><a href="./.github/workflows/lint.yml">lint.yml</a></sub></td>
<td>Runs <strong>golangci-lint</strong> (<a href="./tools/configs/golangci.yaml">config</a>) and <strong>yamllint</strong> (<a href="./tools/configs/yamllint.yaml">config</a>)</td>
<td><code>make deps:lint && make lint</code><br><sub>Auto-fix: <code>make lint:fix</code></sub></td>
</tr>

<tr>
<td>🧹📊 <strong>Lint Then Benchmark</strong><br><sub><a href="./.github/workflows/lint-then-benchmark.yml">lint-then-benchmark.yml</a></sub></td>
<td>Linting + conditional benchmarks<br><sub><code>SHORT</code> for PRs to develop · <code>FULL</code> with label · Skip with <code>action/no-benchmark</code></sub></td>
<td><code>make lint</code><br><code>make test:bench-short</code></td>
</tr>

<tr>
<td>🚀 <strong>Start Binary</strong><br><sub><a href="./.github/workflows/start-binary.yml">start-binary.yml</a></sub></td>
<td>Builds the binary and verifies it starts</td>
<td><code>make build && ./build/defradb start --no-keyring</code></td>
</tr>

<tr>
<td>🧪 <strong>Test Coverage</strong><br><sub><a href="./.github/workflows/test-coverage.yml">test-coverage.yml</a></sub></td>
<td>Comprehensive test matrix: clients (Go/HTTP/CLI), databases, mutations, ACP, lenses, views, encryption, vectors → <a href="https://codecov.io/gh/sourcenetwork/defradb">Codecov</a></td>
<td><code>make test</code> with <a href="#-test-configuration-variables">env variables</a></td>
</tr>

<tr>
<td>🐋 <strong>Validate Containerfile</strong><br><sub><a href="./.github/workflows/validate-containerfile.yml">validate-containerfile.yml</a></sub></td>
<td>Builds Docker image from <a href="./tools/defradb.containerfile">containerfile</a> and verifies it runs</td>
<td>Ensure <a href="./tools/defradb.containerfile">containerfile</a> is valid</td>
</tr>

<tr>
<td>🏷️ <strong>Validate Title</strong><br><sub><a href="./.github/workflows/validate-title.yml">validate-title.yml</a></sub></td>
<td>Validates PR title follows <a href="./CONTRIBUTING.md#-title-format">conventional commit style</a></td>
<td>Fix title per <a href="./CONTRIBUTING.md#-title-format">rules</a> · <a href="./tools/scripts/validate-conventional-style.sh">script</a></td>
</tr>
</table>

### 📊 Additional CI Checks (Non-Blocking)

These also run on PRs but are **informational** - failures won't block merge:

<table>
<tr><th>Check</th><th>What It Does</th></tr>
<tr>
<td>🐢 <strong>Test Limited Resource</strong><br><sub><a href="./.github/workflows/test-limited-resource.yml">test-limited-resource.yml</a></sub></td>
<td>Runs tests on standard (slower) runners to catch resource-constrained failures</td>
</tr>
<tr>
<td>🍎 <strong>Test macOS</strong><br><sub><a href="./.github/workflows/test-macos.yml">test-macos.yml</a></sub></td>
<td>Integration tests on macOS for cross-platform compatibility</td>
</tr>
<tr>
<td>📜 <strong>Test NPX/JS Build</strong><br><sub><a href="./.github/workflows/test-npx.yml">test-npx.yml</a></sub></td>
<td>Verifies NPX/JavaScript-dependent tests can build and run</td>
</tr>
<tr>
<td>☁️ <strong>Preview AMI</strong><br><sub><a href="./.github/workflows/preview-ami-with-terraform-plan.yml">preview-ami-with-terraform-plan.yml</a></sub></td>
<td>Triggers on AWS infra changes only - validates Terraform and posts plan as PR comment</td>
</tr>
</table>

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

<table>
<tr><th>Variable</th><th>Values</th><th>Description</th></tr>
<tr><td><code>DEFRA_CLIENT_GO</code></td><td><code>true</code>/<code>false</code></td><td>Enable Go client tests</td></tr>
<tr><td><code>DEFRA_CLIENT_HTTP</code></td><td><code>true</code>/<code>false</code></td><td>Enable HTTP client tests</td></tr>
<tr><td><code>DEFRA_CLIENT_CLI</code></td><td><code>true</code>/<code>false</code></td><td>Enable CLI client tests</td></tr>
<tr><td><code>DEFRA_BADGER_MEMORY</code></td><td><code>true</code>/<code>false</code></td><td>Use in-memory Badger store</td></tr>
<tr><td><code>DEFRA_BADGER_FILE</code></td><td><code>true</code>/<code>false</code></td><td>Use file-based Badger store</td></tr>
<tr><td><code>DEFRA_BADGER_ENCRYPTION</code></td><td><code>true</code>/<code>false</code></td><td>Enable Badger encryption</td></tr>
<tr><td><code>DEFRA_MUTATION_TYPE</code></td><td><code>gql</code> / <code>collection-named</code> / <code>collection-save</code></td><td>Mutation type</td></tr>
<tr><td><code>DEFRA_DOCUMENT_ACP_TYPE</code></td><td><code>local</code> / <code>source-hub</code></td><td>ACP type</td></tr>
<tr><td><code>DEFRA_LENS_TYPE</code></td><td><code>wasm-time</code> / <code>wasm-er</code></td><td>Lens WASM runtime</td></tr>
<tr><td><code>DEFRA_VIEW_TYPE</code></td><td><code>cacheless</code> / <code>materialized</code></td><td>View type</td></tr>
<tr><td><code>DEFRA_VECTOR_EMBEDDING</code></td><td><code>true</code>/<code>false</code></td><td>Enable vector embedding tests</td></tr>
</table>

### 🌐 SourceHub ACP Tests

> [!WARNING]
> SourceHub ACP tests require **Docker** and are resource-heavy.

```shell
DEFRA_CLIENT_HTTP=true DEFRA_CLIENT_GO=false DEFRA_DOCUMENT_ACP_TYPE=source-hub \
  DEFRA_BADGER_MEMORY=true go test ./tests/integration/acp/... -count=1 -timeout 20m
```

Use `-p 1` when running the full suite to avoid Docker resource contention.

### 📈 Benchmarks

<table>
<tr><td>

**Running benchmarks:**

```shell
make test:bench          # Full benchmark suite
make test:bench-short    # Short benchmark suite
```

</td></tr>
<tr><td>

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

</td></tr>
</table>

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
