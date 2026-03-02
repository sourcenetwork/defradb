<p align="center">
  <img src="docs/assets/contributing-banner.svg" alt="Contributing to DefraDB" width="100%">
</p>

<p align="center">
  <a href="https://discord.gg/w7jYQVJ"><img src="https://img.shields.io/discord/427944769851752448.svg?color=768AD4&label=Discord&logo=discord&logoColor=white" alt="Discord"></a>
  <a href="https://twitter.com/sourcenetwrk"><img src="https://img.shields.io/twitter/follow/sourcenetwrk.svg?label=Follow&style=social" alt="Twitter"></a>
  <a href="https://github.com/sourcenetwork/defradb/blob/develop/licenses/BSL.txt"><img src="https://img.shields.io/badge/license-BSL--1.1-blue" alt="License"></a>
</p>

---

# 🎉 Contributing to DefraDB

**Thank you for your interest in contributing to DefraDB!** You're about to join a wave of innovation in decentralized and powerful databases. Third-party contributions are essential — we simply can't cover every platform, configuration, and use case on our own. Every contribution, no matter how small, makes a difference.

This guide will walk you through everything you need to know — from reporting bugs and suggesting features to writing code and getting your pull request merged.

> 💡 **New here?** The quickest way to get started is to join our [Discord community](https://discord.gg/w7jYQVJ) and say hello in the **#general** channel. We're happy to help you find the right place to contribute!

---

## 📑 Table of Contents

- [🎉 Contributing to DefraDB](#-contributing-to-defradb)
  - [📑 Table of Contents](#-table-of-contents)
  - [📋 Prerequisites](#-prerequisites)
  - [🔐 Security Vulnerabilities](#-security-vulnerabilities)
  - [🚀 Getting Started](#-getting-started)
    - [🛠️ Build Prerequisites](#️-build-prerequisites)
    - [📥 Clone, Build, and Run](#-clone-build-and-run)
  - [📖 Documentation](#-documentation)
    - [📄 Man Pages](#-man-pages)
  - [🐛 Reporting Bugs](#-reporting-bugs)
  - [💡 Suggesting Enhancements](#-suggesting-enhancements)
  - [🌊 Git Workflow](#-git-workflow)
    - [🍴 Fork Flow (All Contributors)](#-fork-flow-all-contributors)
    - [🔀 Limited Use of Branch Flow (Internal Only)](#-limited-use-of-branch-flow-internal-only)
  - [📬 Opening a Pull Request](#-opening-a-pull-request)
    - [🔗 Link with Relevant Issue(s)](#-link-with-relevant-issues)
    - [🏷️ Title Format](#️-title-format)
    - [✍️ Sign the CLA](#️-sign-the-cla)
  - [👀 Managing Pull Requests](#-managing-pull-requests)
    - [🙋 Asking for Review](#-asking-for-review)
    - [💬 Commenting Etiquette](#-commenting-etiquette)
  - [⚙️ CI Checks — Pass Before Merge](#️-ci-checks--pass-before-merge)
    - [✅ Required CI Checks](#-required-ci-checks)
    - [📊 Additional CI Checks](#-additional-ci-checks)
  - [🏁 Merging Pull Requests](#-merging-pull-requests)
    - [🟢 Ready to Merge Checklist](#-ready-to-merge-checklist)
  - [🧪 Testing](#-testing)
    - [🏃 Running Tests Locally](#-running-tests-locally)
    - [📈 Benchmarks](#-benchmarks)
  - [📏 Code Style and Quality](#-code-style-and-quality)
  - [📦 Dependency Management](#-dependency-management)
    - [🐹 Go Version Bumping Policy](#-go-version-bumping-policy)
  - [📝 Additional Information](#-additional-information)
    - [💡 Ideation / Proposals](#-ideation--proposals)
    - [📋 Project Management](#-project-management)
    - [⚖️ Licensing](#️-licensing)
    - [📣 Community](#-community)

---

## 📋 Prerequisites

Before you begin, familiarize yourself with the following technologies and resources:

| Resource | Why You Need It |
|----------|----------------|
| 📘 [Project Documentation](https://docs.source.network/) | Understand DefraDB's features, architecture, and query language |
| 🔀 [Git](https://training.github.com/) | Version control fundamentals — branching, rebasing, commits |
| 🐙 [GitHub](https://docs.github.com/) | Pull requests, issues, and collaboration workflows |
| 🐹 [Go](https://go.dev/doc/install) | Primary language — install the Go toolchain |
| 🦀 [Cargo/rustc](https://doc.rust-lang.org/cargo/commands/cargo-rustc.html) via [rustup](https://www.rust-lang.org/tools/install) | Required for building WebAssembly lens modules |
| 🌐 [SourceHub](https://github.com/sourcenetwork/sourcehub) | The SourceHub network for access control features |
| 🤖 [Ollama](https://ollama.com/download) | Required for AI/vector embedding tests |

---

## 🔐 Security Vulnerabilities

> ⚠️ **Please do NOT file a public issue for security vulnerabilities.**

If you discover a security vulnerability, please disclose it responsibly by emailing **[security@source.network](mailto:security@source.network)**. Our security team will respond within 24 hours.

For full details, see our [Security Policy](./SECURITY.md).

---

## 🚀 Getting Started

### 🛠️ Build Prerequisites

Make sure you have the following installed on your system:

- **[Go](https://go.dev/doc/install)** — check our [`go.mod`](./go.mod) for the required version
- **[Cargo/rustc](https://www.rust-lang.org/tools/install)** — via `rustup`, needed for WASM lens modules
- **[SourceHub](https://github.com/sourcenetwork/sourcehub)** — install via `make install`
- **[Ollama](https://ollama.com/download)** — required for vector embedding tests (install and run with `make deps:ollama && make ollama`)
- **[Make](https://www.gnu.org/software/make/)** — build automation tool

> 💡 **Tip:** Run `make deps` to install all project dependencies at once.

### 📥 Clone, Build, and Run

```shell
git clone https://github.com/sourcenetwork/defradb.git
cd defradb
make start
```

Refer to the [`README.md`](./README.md) and the [project documentation](https://docs.source.network/) for detailed usage examples including schema definitions, queries, peer-to-peer synchronization, and more.

---

## 📖 Documentation

The overall project documentation can be found at **[docs.source.network](https://docs.source.network)**, and its source at [github.com/sourcenetwork/docs.source.network](https://github.com/sourcenetwork/docs.source.network).

**Code documentation (Go doc comments)** can be viewed as a website:

```shell
go install golang.org/x/pkgsite/cmd/pkgsite@latest
cd your-path-to/defradb/
pkgsite
# open http://localhost:8080/github.com/sourcenetwork/defradb
```

- 📝 Refer to [go.dev/doc/comment](https://go.dev/doc/comment) for guidelines on writing Go doc comments.
- 🌐 The [`docs/website/references/http/openapi.json`](./docs/website/references/http/openapi.json) file contains auto-generated HTTP API documentation.
- 💻 The [`docs/website/references/cli/`](./docs/website/references/cli/) directory contains auto-generated CLI documentation.

### 📄 Man Pages

We support man pages. To generate them:

```sh
make docs:manpages
```

The man pages will be placed in the `build/man/` directory. You can install them manually by copying to your system's man directory, or on Linux:

```sh
make install:manpages
```

---

## 🐛 Reporting Bugs

Found a bug? Here's how to report it effectively:

1. **Search first** — Check [existing issues](https://github.com/sourcenetwork/defradb/issues) to avoid duplicates.
2. **Create a new issue** — Go to [github.com/sourcenetwork/defradb/issues](https://github.com/sourcenetwork/defradb/issues), click "New issue", and select **Bug Report**.
3. **Fill out the template** with:
   - 📝 A clear and concise description of the bug
   - 🔄 Steps to reproduce the behavior
   - ✅ Expected behavior vs actual behavior
   - 💻 Platform info (use `defradb version` to get version details)
   - 📎 Any additional context (logs, screenshots, etc.)

> 💡 **Pro tip:** The more information you provide, the faster we can fix it. Include relevant log output and the DefraDB version.

---

## 💡 Suggesting Enhancements

Have an idea for a new feature or improvement?

1. **Start a discussion** — Open a [Feature Request discussion](https://github.com/sourcenetwork/defradb/discussions/new?category=feature-request) to share your idea.
2. **Provide context** — Explain the problem you're trying to solve and why this enhancement would be useful.
3. **Consider scope** — For large architectural changes, consider writing a [Source Improvement Proposal (SIP)](https://github.com/sourcenetwork/SIPs/).

> 💡 **Tip:** For significant changes, it's a good idea to discuss your approach with the team *before* writing code. This prevents wasted effort on approaches that might not align with the project direction.

---

## 🌊 Git Workflow

We have adopted the **Git Fork Flow** as our primary development workflow for both internal and external contributors. This ensures a consistent and streamlined approach to contributions across the project.

<p align="center">
  <img src="docs/assets/fork-flow.svg" alt="Fork Flow Diagram" width="100%">
</p>

### 🍴 Fork Flow (All Contributors)

All developers, whether internal or external, are expected to use the Fork Flow:

1. 🍴 **Fork** the main repository on GitHub.
2. 📥 **Clone** your forked repository locally:
   ```shell
   git clone https://github.com/<your-username>/defradb.git
   cd defradb
   ```
3. 🔗 **Add upstream** remote to keep your fork in sync:
   ```shell
   git remote add upstream https://github.com/sourcenetwork/defradb.git
   ```
4. 🌿 **Create a feature branch** from an up-to-date `develop`:
   ```shell
   git fetch upstream
   git checkout -b your-feature-branch upstream/develop
   ```
5. 💻 **Make your changes** on the feature branch.
6. 🧪 **Write tests** for any modified behavior, if applicable.
7. ✅ **Run all required local checks** to ensure everything passes:
   ```shell
   make docs          # Regenerate documentation
   make mocks         # Regenerate mocks
   make lint          # Run linters
   make tidy          # Run go mod tidy
   make test          # Run unit and integration tests
   make test:changes  # Run data format change detector
   ```
8. 📝 **Commit** your changes to the feature branch.
9. 🚀 **Push** your feature branch to your fork:
   ```shell
   git push origin your-feature-branch
   ```
10. 📬 **[Open a pull request](#-opening-a-pull-request)** from your branch on your fork, targeting the **`develop`** branch of the main repository.

### 🔀 Limited Use of Branch Flow (Internal Only)

In specific cases, internal developers may need to use the Branch Flow, especially for CI-related updates that require direct access to the main repository. If an internal developer opts for this approach:

1. Clone the main repository locally.
2. Create a feature branch following the **branch naming convention**:
   ```
   <dev-name>/<label>/<description>
   ```
   **Example:** `lone/ci/update-test-workflow-action`

3. Make changes on the feature branch.
4. Write tests for any modified behavior, if applicable.
5. Run all required local checks:
   ```shell
   make docs && make mocks && make lint && make tidy && make test
   ```
6. Commit, push, and [open a pull request](#-opening-a-pull-request) targeting the `develop` branch.

> ⚠️ **Note:** The use of Branch Flow should be limited even amongst internal developers. Fork Flow is always preferred.

---

## 📬 Opening a Pull Request

<p align="center">
  <img src="docs/assets/pr-lifecycle.svg" alt="PR Lifecycle" width="100%">
</p>

### 🔗 Link with Relevant Issue(s)

Since we follow **issue-driven development**, every pull request **must** be [linked](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/using-keywords-in-issues-and-pull-requests#linking-a-pull-request-to-an-issue) to one or more issues. If no issue exists for your change, create one first.

Link an issue by adding the issue number in your PR description using any of the [resolving keywords](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/using-keywords-in-issues-and-pull-requests#linking-a-pull-request-to-an-issue):

| Keyword | Example |
|---------|---------|
| `close` / `closes` / `closed` | `Closes #123` |
| `fix` / `fixes` / `fixed` | `Fixes #456` |
| `resolve` / `resolves` / `resolved` | `Resolves #789` |

**Example PR description:**

```markdown
## Relevant issue(s)

Resolves #123
```

> 📝 **Note:** There is a `Relevant issue(s)` section at the top of the [PR description template](./.github/pull_request_template.md) just for this purpose.

### 🏷️ Title Format

Every pull request title **must** follow a structured format inspired by the **[Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)** style:

```
<label>: <Description>
```

**Available labels:**

| Label | Meaning | Example |
|-------|---------|---------|
| `feat` | ✨ New feature | `feat: Add document encryption` |
| `fix` | 🐛 Bug fix | `fix: Resolve API timeout issue` |
| `refactor` | ♻️ Code refactoring | `refactor: Simplify query parser` |
| `docs` | 📖 Documentation | `docs: Update API reference` |
| `test` | 🧪 Testing | `test: Add P2P sync tests` |
| `ci` | 🔧 CI/CD changes | `ci: Update lint workflow` |
| `chore` | 🧹 Maintenance | `chore: Update dependencies` |
| `perf` | ⚡ Performance | `perf: Optimize index lookups` |
| `tools` | 🔨 Tooling | `tools: Add benchstat script` |
| `bot` | 🤖 Automated (bot only) | `bot: Bump go version` |

**Title Rules:**

| # | Rule | Example |
|---|------|---------|
| 1️⃣ | A colon (`:`) must follow `<label>`, with **a single space** after it | ✅ `fix: Resolve issue` &nbsp;&nbsp; ❌ `fix:Resolve issue` |
| 2️⃣ | `<Description>` must start with **a capital letter** | ✅ `docs: Improve README` &nbsp;&nbsp; ❌ `docs: improve README` |
| 3️⃣ | `<Description>` **should** begin with an **action verb** | See [suggested verbs](https://gist.github.com/scmx/411f6fea4ee3832806720d536a7d5d8f) |
| 4️⃣ | Last character **must** be alphanumeric (`a-z`, `A-Z`, `0-9`) | ❌ No trailing periods or special chars |
| 5️⃣ | If label is **not** `bot`, title must **not exceed 60 characters** | Keep it concise! |

> 📌 More examples (valid/invalid PR titles) can be found in [`tools/scripts/scripts_test.sh`](./tools/scripts/scripts_test.sh).

### ✍️ Sign the CLA

First-time contributors will be asked to read and accept the **Contributor License Agreement (CLA)**. The CLA bot will automatically comment on your PR with instructions.

---

## 👀 Managing Pull Requests

### 🙋 Asking for Review

- Request a review from the **database-team**.
- Discuss and adapt the pull request as needed by following the commenting etiquette below.

> 💡 **Response expectations:** After receiving feedback, please try to respond within **two weeks**. This keeps the conversation active and helps us merge contributions in a timely manner.

### 💬 Commenting Etiquette

It can sometimes be unclear whether a reviewer's comment is blocking or non-blocking. To address this, we've adopted labels inspired by **[Conventional Comments](https://conventionalcomments.org/)** to clarify the nature and urgency of each comment.

| Label | Icon | Meaning | Action Required |
|-------|------|---------|----------------|
| `thought` | 💭 | A dump of thoughts — may or may not be within scope. Provides context or sparks ideas. | No action required. |
| `question` | ❓ | A question from the reviewer. May evolve into other types once clarity is achieved. | Answer the question, or point the reviewer to a helpful resource. |
| `nitpick` | 🔍 | Minor, nitpicky suggestion. | Can be ignored or accepted — no follow-up required. |
| `suggestion` | 💡 | Non-blocking suggestion. | Accept it, or explain why it shouldn't be done. |
| `todo` | 📋 | **Blocking** — must be resolved before merge. | Must resolve before merge. If deferring to another PR, create an issue and link it. |
| `issue` | 🚨 | **Major blocking issue** — MUST be resolved in this PR. | Must resolve in this PR. Cannot be deferred. |

**Example review comment:**

```
suggestion: Consider using a map here instead of a slice for O(1) lookups.
```

---

## ⚙️ CI Checks — Pass Before Merge

When you open a PR, our CI pipeline will automatically run a comprehensive suite of checks. **All required checks must pass** before your PR can be merged.

### ✅ Required CI Checks

These checks run on every PR targeting `develop` or `master`:

| Check | Workflow | What It Does | How to Fix Locally |
|-------|----------|-------------|-------------------|
| 🏗️ **Build Dependencies** | [`build-dependencies.yml`](./.github/workflows/build-dependencies.yml) | Ensures all project dependencies can be built successfully. | Run `make deps` |
| 📊 **Check Data Format Changes** | [`check-data-format-changes.yml`](./.github/workflows/check-data-format-changes.yml) | Detects backwards-incompatible data format changes. If changes are detected, they must be documented in [`docs/data_format_changes/`](./docs/data_format_changes/README.md). | Run `make test:changes`. Learn more about the [change detector](./tests/change_detector/README.md). |
| 📖 **Check Documentation** | [`check-documentation.yml`](./.github/workflows/check-documentation.yml) | Ensures CLI docs, HTTP API docs, and README table of contents are up to date. Runs three sub-checks: `check-cli-documentation`, `check-http-documentation`, and `check-readme-toc`. | Run `make docs` (which runs `make docs:cli`, `make docs:http`, and `make toc`). |
| 🔧 **Check Mocks** | [`check-mocks.yml`](./.github/workflows/check-mocks.yml) | Verifies that all mocks are regenerated and up to date. | Run `make mocks` |
| 📦 **Check Tidy** | [`check-tidy.yml`](./.github/workflows/check-tidy.yml) | Ensures `go.mod` and `go.sum` are in a clean state. | Run `make tidy` |
| 🔒 **Check Vulnerabilities** | [`check-vulnerabilities.yml`](./.github/workflows/check-vulnerabilities.yml) | Runs `govulncheck` to scan for known security vulnerabilities in dependencies. | Run `make deps:vulncheck && govulncheck ./...` |
| 🧙 **Check Wizard Health** | [`check-wizard-health.yml`](./.github/workflows/check-wizard-health.yml) | Tests the interactive setup wizard using an automated expect script to ensure it works correctly. | Run `make build && ./tools/scripts/wizard_test.sh` |
| 🧹 **Lint** | [`lint.yml`](./.github/workflows/lint.yml) | Runs **golangci-lint** for Go code and **yamllint** for YAML files. | Run `make deps:lint` to install, then `make lint`. Use `make lint:fix` for auto-fixable issues. |
| 🧹📊 **Lint Then Benchmark** | [`lint-then-benchmark.yml`](./.github/workflows/lint-then-benchmark.yml) | Runs linting + conditional benchmarks. Benchmarks are `SHORT` for PRs to develop, `FULL` with the `action/full-benchmark` label, or skipped with `action/no-benchmark`. | Lint: `make lint`. Benchmark: `make test:bench` or `make test:bench-short`. |
| 🚀 **Start Binary** | [`start-binary.yml`](./.github/workflows/start-binary.yml) | Builds the binary and verifies it can actually start. | Run `make build && ./build/defradb start --no-keyring` |
| 🧪 **Test Coverage** | [`test-coverage.yml`](./.github/workflows/test-coverage.yml) | Runs the comprehensive test matrix across multiple client types (Go/HTTP/CLI), databases (file/memory), mutation types, ACP modes, lens types, views, encryption modes, and vector embeddings. Uploads coverage to [Codecov](https://codecov.io/gh/sourcenetwork/defradb). | Run `make test`. Use additional `ENV` variables for specific test configurations — see the [workflow file](./.github/workflows/test-coverage.yml) for details. |
| 🐋 **Validate Containerfile** | [`validate-containerfile.yml`](./.github/workflows/validate-containerfile.yml) | Builds a Docker image from the [containerfile](./tools/defradb.containerfile) and verifies it runs correctly. | Ensure the [containerfile](./tools/defradb.containerfile) is valid. |
| 🏷️ **Validate Title** | [`validate-title.yml`](./.github/workflows/validate-title.yml) | Validates the PR title follows the [conventional commit style](#️-title-format). Runs automatically when the title changes. | Fix your PR title to match the [format rules](#️-title-format). See the [validation script](./tools/scripts/validate-conventional-style.sh). |

### 📊 Additional CI Checks

These checks also run on PRs but are **not blocking** (informational or conditional):

| Check | Workflow | What It Does |
|-------|----------|-------------|
| 🐢 **Test Limited Resource** | [`test-limited-resource.yml`](./.github/workflows/test-limited-resource.yml) | Runs tests on standard (slower) GitHub runners to detect failures specific to resource-constrained environments. Not required to pass. |
| 🍎 **Test macOS** | [`test-macos.yml`](./.github/workflows/test-macos.yml) | Runs integration tests on macOS to ensure cross-platform compatibility. |
| 📜 **Test NPX/JS Build** | [`test-npx.yml`](./.github/workflows/test-npx.yml) | Verifies tests that depend on NPX/JavaScript can build and run. |
| ☁️ **Preview AMI** | [`preview-ami-with-terraform-plan.yml`](./.github/workflows/preview-ami-with-terraform-plan.yml) | Only triggers when AWS infrastructure files are changed. Validates Terraform configuration and posts a plan preview as a PR comment. |

> 💡 **Quick fix checklist** — If CI is failing, try these commands locally:
> ```shell
> make tidy          # Fix go.mod/go.sum issues
> make docs          # Regenerate documentation
> make mocks         # Regenerate mocks
> make lint          # Check for lint errors
> make lint:fix      # Auto-fix lint errors where possible
> make test          # Run the full test suite
> make test:changes  # Check for data format changes
> ```

---

## 🏁 Merging Pull Requests

### 🟢 Ready to Merge Checklist

A PR is ready to merge when **all** of the following are true:

- [x] ✅ All required CI checks are passing
- [x] 👍 At least **one approval** from a reviewer
- [x] 🚫 No `DO NOT MERGE` label on the PR
- [x] 🔄 Rebased with the upstream `develop` branch

> 📝 We follow the **[Squash and Merge](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/incorporating-changes-from-a-pull-request/about-pull-request-merges)** strategy. All commits in your PR will be squashed into a single commit on merge. Make sure your PR title follows the [conventional commit style](#️-title-format) as it becomes the final commit message.

**For internal developers:**
Click "Squash and merge" to combine the commits and merge into the `develop` branch. Ensure the final commit title matches the [conventional commit format](#️-title-format).

**For external contributors:**
Inform a maintainer that your PR is ready to merge, and they will handle the merge for you.

> ⚠️ **Breaking changes:** When introducing breaking changes, include the `BREAKING CHANGE` keyword in the commit message body (not the title), followed by a description of the change. This helps track changes that may require additional attention or migration steps.

---

## 🧪 Testing

### 🏃 Running Tests Locally

DefraDB has a comprehensive test suite. Here are the key commands:

| Command | What It Does |
|---------|-------------|
| `make test` | Run unit and integration tests |
| `make test:quick` | Run a quick subset of tests |
| `make test:changes` | Run the data format change detector |
| `make test:build` | Verify the test suite builds |
| `make test:names` | Run tests with named output (useful for CI) |
| `make test:coverage` | Run tests with coverage reporting |
| `make test:coverage-html` | Generate an HTML coverage report |

**Environment variables** control which test configurations run:

| Variable | Values | Description |
|----------|--------|-------------|
| `DEFRA_CLIENT_GO` | `true`/`false` | Enable Go client tests |
| `DEFRA_CLIENT_HTTP` | `true`/`false` | Enable HTTP client tests |
| `DEFRA_CLIENT_CLI` | `true`/`false` | Enable CLI client tests |
| `DEFRA_BADGER_MEMORY` | `true`/`false` | Use in-memory Badger store |
| `DEFRA_BADGER_FILE` | `true`/`false` | Use file-based Badger store |
| `DEFRA_BADGER_ENCRYPTION` | `true`/`false` | Enable Badger encryption |
| `DEFRA_MUTATION_TYPE` | `gql`/`collection-named`/`collection-save` | Mutation type |
| `DEFRA_DOCUMENT_ACP_TYPE` | `local`/`source-hub` | ACP type |
| `DEFRA_LENS_TYPE` | `wasm-time`/`wasm-er` | Lens WASM runtime |
| `DEFRA_VIEW_TYPE` | `cacheless`/`materialized` | View type |
| `DEFRA_VECTOR_EMBEDDING` | `true`/`false` | Enable vector embedding tests |

> 💡 **Tip:** For SourceHub ACP tests, you'll need Docker running as they spin up SourceHub containers. Run them with:
> ```shell
> DEFRA_CLIENT_HTTP=true DEFRA_CLIENT_GO=false DEFRA_DOCUMENT_ACP_TYPE=source-hub \
>   DEFRA_BADGER_MEMORY=true go test ./tests/integration/acp/... -count=1 -timeout 20m
> ```

### 📈 Benchmarks

Run the benchmark suite to measure performance:

```shell
make test:bench          # Full benchmark suite
make test:bench-short    # Short benchmark suite
```

To compare your branch's performance against `develop`:

```shell
# On develop branch
make test:bench | tee develop.txt

# On your feature branch
make test:bench | tee current.txt

# Compare results
make deps:bench     # Install benchstat
benchstat develop.txt current.txt
```

> 📝 **CI benchmarks:** The [Lint Then Benchmark workflow](./.github/workflows/lint-then-benchmark.yml) automatically runs benchmarks on PRs to `develop`. Use the `action/full-benchmark` label for a full benchmark run, or `action/no-benchmark` to skip.

---

## 📏 Code Style and Quality

We strive for clean, consistent code. Here are the key guidelines:

- 🐹 **Go formatting** — All Go code must be formatted. The linter enforces this.
- 📝 **Comments** — Refer to [go.dev/doc/comment](https://go.dev/doc/comment) for Go doc comment guidelines.
- ⚖️ **License header** — Include the [BSL license header](./licenses/BSL.txt) at the top of every new code file.
- 🧹 **Linting** — We use [golangci-lint](https://golangci-lint.run/) with a [custom configuration](./tools/configs/golangci.yaml) and [yamllint](https://yamllint.readthedocs.io/) with its own [config](./tools/configs/yamllint.yaml).
- 🧪 **Testing** — Write tests for new features and bug fixes. Tests should pass at every commit in your PR.

```shell
make deps:lint    # Install linting tools
make lint         # Run all linters
make lint:fix     # Auto-fix where possible
```

---

## 📦 Dependency Management

- Run `make tidy` for any PR that changes dependencies. This runs `go mod tidy` to keep `go.mod` and `go.sum` clean.
- The project uses **Dependabot** to automatically create PRs for dependency updates. These are periodically combined using the [Combine Bot PRs workflow](./.github/workflows/combine-bot-prs.yml).
- When adding a new dependency, evaluate its maintenance status, security track record, and license compatibility.

### 🐹 Go Version Bumping Policy

Our Go version bumping policy ensures stability while staying current:

1. **🎯 One version behind** — We use **one version behind** the latest Go release. A Go release becomes unsupported when the second new major version is released after it. The upper limit prevents bleeding-edge instabilities.

2. **🔒 Security exceptions** — If `govulncheck` reports vulnerabilities fixed only in the latest Go version (and patches haven't landed on our current version within ~24 hours), we **do not** bump preemptively.
   See the [vulnerability check workflow](./.github/workflows/check-vulnerabilities.yml).

3. **📦 Dependency-driven bumps** — If a dependency strictly requires a newer Go version and DefraDB can't resolve the issue otherwise, we bump accordingly.

---

## 📝 Additional Information

### 💡 Ideation / Proposals

The community follows the **[Source Improvement Proposals (SIPs)](https://github.com/sourcenetwork/SIPs/)** process for comprehensive changes. If you're planning a significant architectural change or new feature, consider writing a SIP first to get community feedback.

### 📋 Project Management

- Use [Milestones](https://github.com/sourcenetwork/defradb/milestones) and the [project board](https://github.com/orgs/sourcenetwork/projects/3) to coordinate work on releases.
- Check the project board for tasks labeled `good-first-issue` if you're looking for a place to start.

### ⚖️ Licensing

- Include the [BSL license header](./licenses/BSL.txt) at the top of **every** code file.
- DefraDB is released under the [Business Source License (BSL)](./licenses/BSL.txt). Each dated version converts to Apache License v2.0 after four years.

### 📣 Community

We'd love to hear from you! Connect with us:

| Channel | Link | Purpose |
|---------|------|---------|
| 💬 **Discord** | [discord.gg/w7jYQVJ](https://discord.gg/w7jYQVJ) | Real-time chat, Q&A, and community discussions |
| 🐦 **Twitter** | [@sourcenetwrk](https://twitter.com/sourcenetwrk) | Project updates and announcements |
| 💻 **GitHub Discussions** | [Discussions](https://github.com/sourcenetwork/defradb/discussions) | Feature requests, ideas, and longer-form conversations |

---

<p align="center">
  <b>🙏 Thank you for contributing to DefraDB! Together, we're building the future of decentralized databases. 🚀</b>
</p>
