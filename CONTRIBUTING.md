## Contributing to DefraDB

Thank you for your interest in contributing to DefraDB! You're about to join a wave of innovation in decentralized and powerful databases. This guide will help you navigate the contribution process, whether you're identifying issues, suggesting features, or contributing code.

Before you begin, it is recommended that you familiarize yourself with the following:

- [Project documentation](https://docs.source.network/)
- [Git](https://training.github.com/)
- [GitHub](https://docs.github.com/)
- [Go](https://go.dev/doc/install)
- [Cargo/rustc](https://doc.rust-lang.org/cargo/commands/cargo-rustc.html), typically installed via [rustup](https://www.rust-lang.org/tools/install)
- [SourceHub](https://github.com/sourcenetwork/sourcehub)
- [Ollama](https://ollama.com/download)

We also encourage you to join the [Source Network Discord](https://discord.gg/w7jYQVJ) to discuss ideas, ask questions, and find inspiration for future developments.

## Documentation

The overall project documentation can be found at [docs.source.network](https://docs.source.network), and its source at [github.com/sourcenetwork/docs.source.network](https://github.com/sourcenetwork/docs.source.network).

To view the code documentation (doc comments) as a website, follow these steps:

```shell
go install golang.org/x/pkgsite/cmd/pkgsite@latest
cd your-path-to/defradb/
pkgsite
# open http://localhost:8080/github.com/sourcenetwork/defradb
```

- Refer to [go.dev/doc/comment](https://go.dev/doc/comment) for guidelines on writing Go doc comments.
- The `docs/website/references/http/openapi.json/` directory contains auto-generated HTTP API documentation.
- The `docs/website/references/cli/` directory contains auto-generated CLI documentation.

### Manpages

We support man pages. To generate them, run:

```sh
make docs:manpages
```

This will place the man pages in the `build/man/` directory.
To install them manually, copy the files to your system's man directory. If you're on Linux, you can install them with:

```sh
make install:manpages
```

## Getting started

To get started, clone the repository, build, and run it:

```shell
git clone https://github.com/sourcenetwork/defradb.git
cd defradb
make start
```

Refer to the [`README.md`](./README.md) and project documentation for usage examples.

## Creating an issue

1. Go to [github.com/sourcenetwork/defradb/issues](https://github.com/sourcenetwork/defradb/issues), and click "New issue".
2. Select the relevant issue type.
3. Fill out the issue template.

## Git workflow

### Fork Flow vs Branch Flow

We have adopted the Git Fork Flow as our primary development workflow for both internal and external contributors. This ensures a consistent and streamlined approach to contributions across the project.

### Fork flow for all contributors

All developers, both internal and external, are required to follow the Git Fork Flow for their contributions. This process includes:

1. Fork the main repository.
2. Clone your forked repository locally.
3. Create a feature branch in your cloned forked repository.
4. Make changes on the feature branch.
5. Write tests for any modified behavior, if applicable.
6. Run required checks to ensure all `make` targets pass:
   - `make docs`
   - `make mocks`
   - `make lint`
   - `make tidy`
   - `make test`
   - `make test:changes`

7. Commit your changes to the feature branch.
8. Push your feature branch with the committed changes to your fork.
9. [Open a pull request](#opening-pull-request) from your branch on your fork, targeting the `develop` branch of the main repository.

This ensures a structured, high-quality, and consistent development process.

### Limited use of Branch Flow

In specific cases, internal developers may need to use the Branch Flow, especially for CI-related updates that require direct access to the main repository. If an internal developer opts for this approach, the process is as follows:

1. Clone the main repository locally.
2. Create a feature branch in the cloned directory, following the branch naming convention:

    ```
     <dev-name>/<label>/<description>
    ```

    Example: `lone/ci/update-test-workflow-action`

3. Make changes on the feature branch.
4. Write tests for any modified behavior, if applicable.
5. Run required checks to ensure all `make` targets pass:
    - `make docs`
    - `make mocks`
    - `make lint`
    - `make tidy`
    - `make test`
6. Commit your changes to the feature branch.
7. Push your feature branch with the committed changes to the main repository.
8. [Open a pull request](#opening-pull-request) on the main repository, targeting the `develop` branch.

**Note:** The use of Branch Flow should be limited even amongst internal developers.

## Opening Pull Request

### Link with relevant issue(s)

Since we follow **Issue-driven development**, every pull request **must** be [linked](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/using-keywords-in-issues-and-pull-requests#linking-a-pull-request-to-an-issue) to one or more issues. To link an issue, include the issue number in your pull request description using any of the [resolving keywords](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/using-keywords-in-issues-and-pull-requests#linking-a-pull-request-to-an-issue) such as:

- close
- closes
- closed
- fix
- fixes
- fixed
- resolve
- resolves
- resolved

This ensures proper tracking and accountability for all contributions.

For example, this PR links to issue `#123`:

```
## Relevant issue(s)
Resolves #123
```

**Note**: There is a `Relevant issue(s)` section at the top of the PR description template just for this purpose.

### Title format

Every pull request title must follow a structured format using a leading label from a fixed set, inspired by the [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) style.

The pull request title must be in this format `<label>: <Description>` where:

- `<label>` is one of the predefined labels:
  - `chore` – Routine or maintenance task.
  - `ci` – CI/CD-related task.
  - `docs` – Documentation-related task.
  - `feat` – Adding a new feature.
  - `fix` – Fixing an issue.
  - `perf` – Performance improvement.
  - `refactor` – Code refactoring.
  - `test` – Testing-related work.
  - `tools` – Tooling-related work.
  - `bot` – Automated task (should not be used manually).

#### Rules

1. A colon (`:`) must follow `<label>`, with **a single space** after it.
   - ✅ **Valid:** `fix: Resolve API timeout issue`
   - ❌ **Invalid (missing space):** `fix:invalid example`

1. `<Description>` must start with **a capital letter**.
   - ✅ **Valid:** `docs: Improve README instructions`
   - ❌ **Invalid (lowercase start):** `docs: improve README instructions`

1. `<Description>` **should** begin with an **action verb** (e.g., "Add", "Update", "Fix").
   - 🔗 [Suggested action verbs](https://gist.github.com/scmx/411f6fea4ee3832806720d536a7d5d8f)

1. The **last character** must be **alphanumeric** (`a-z`, `A-Z`, `0-9`).

1. If the label is **not** `bot`, the **entire title** must **not exceed 60 characters**.

📌 More examples (valid/invalid PR titles) can be found in [tools/scripts/scripts_test.sh](https://github.com/sourcenetwork/defradb/blob/develop/tools/scripts/scripts_test.sh).

### Sign the CLA

Read and accept the Contributor License Agreement (for first-time contributors).

## Managing PRs (pull requests)

### Asking for review

- Request a review from the *database-team*.
- Discuss and adapt the pull request as necessary, following the commenting etiquette outlined below.

### Commenting etiquette

It can sometimes be unclear whether a reviewer's comment is blocking or non-blocking. To address this, we've adopted labels inspired by **[conventional comments](https://conventionalcomments.org/)** to help clarify the nature of the comments.

| Label      | Meaning                                                                                         | Action                                                                                          |
|------------|-------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------|
| `thought`  | These comments are a dump of thoughts. They may or may not be within the scope of this PR. They can provide context or just spark ideas. | No action is required.                                                                         |
| `question` | These comments are questions the reviewer has. They may evolve into other types once clarity is achieved. | Answer the question or guide the reviewer to a resource that addresses it.                      |
| `nitpick`  | Minor, nitpicky suggestions that the reviewer might have.                                         | Can be ignored or accepted, with no follow-up required.                                          |
| `suggestion` | Non-blocking suggestions that the reviewer might have.                                          | Either accept the suggestion or provide context on why it shouldn't be done.                    |
| `todo`     | Blocking suggestions that must be resolved before merging.                                        | Must resolve these before the PR can be merged. If resolving in another PR, create an issue and link it. |
| `issue`    | Major blocking issue that **MUST** be resolved in this PR.                                        | Must resolve this in the current PR.                                                            |

## Merging PRs

### Pass CI checks

These are the required CI checks, that **MUST** pass in order to have the PR merged.

| Check Type             | Description                                                                                                                      | Resolve                                                                                                                                                                         |
|------------------------|----------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `Check Data Format Changes` | Ensures no data format changes occurred. If they did occur, resolve by documenting the changes in this directory: `docs/data_format_changes/`. More instructions [here](./docs/data_format_changes/README.md). | To run the change detector, simply do `make test:changes`. More about how it works [here](./tests/change_detector/README.md).                                                   |
| `Check Documentation`   | Ensures all documentation is up to date. Some documentation might need to be re-generated depending on the changes.              | To generate the documentation, do `make docs`.                                                                                                                                 |
| `Check Lint`            | Ensures all linting rules are adhered to.                                                                                         | To install the linting tools, do `make deps:lint`. To see the lint failures, do `make lint`. In some cases, to auto-fix linter failures, do `make lint:fix`.                    |
| `Check Mocks`           | Ensures all mocks are up to date. Some mocks might need to be regenerated depending on the changes.                              | To generate the mocks, do `make mocks`.                                                                                                                                       |
| `Check Tidy`            | Ensures `go mod tidy` is in a clean state.                                                                                        | Run `make tidy`.                                                                                                                                                              |
| `Start Binary`          | Ensures the binary can be started after building.                                                                                | To build, do `make build`, then start with `./build/defradb start --no-keyring`.                                                                                              |
| `Test Coverage`         | Ensures the different combinations of tests work (also generates coverage).                                                     | To run the tests, do `make test`. However, you may need to use additional `ENV` variables to trigger specific kinds of tests. For more details on each test job, see the [workflow file](.github/workflows/test-coverage.yml). |
| `Validate Container`    | Ensures the container file did not break.                                                                                        | Resolve by ensuring the [container file](./tools/defradb.containerfile) is still valid.                                                                                        |
| `Validate Title`        | Ensures the PR title adheres to the rules described [above](#title-format).                                                      | This runs automatically when the title changes. See how it works [here](./tools/scripts/validate-conventional-style.sh).                                                        |

### PR ready to merge

A PR is ready to merge when:

- There is at least one approval.
- There is no `DO NOT MERGE` label.
- It is rebased with the upstream `develop` branch.
- All required checks are passing

Note: We follow the **[Squash and Merge](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/incorporating-changes-from-a-pull-request/about-pull-request-merges)** strategy.

#### *Internal developer(s)*

Click the "Squash and merge" to combine the commits and merge into the `develop` branch.

#### *External contributor(s)*

Inform a maintainer that you are ready to have to PR merged, and the maintainer will merge the PR for you.

## Extras

### Version bumping policy

1. **Go version**:  
   We follow a policy of using **one version behind** the latest Go version. This ensures that we are not on the bleeding edge of Go releases while still maintaining support. The lower limit is set because Go releases become unsupported as soon as the second new major release is cut. The upper limit is to avoid potential instability from the latest Go version.

2. **Security vulnerabilities**:  
   If `govulncheck` reports vulnerabilities fixed in the latest Go version, but those fixes have not yet been applied to the patch of the Go version we are currently using (usually within 24 hours of a known vulnerability), we do **not** bump our Go version.  
   [Check vulnerabilities action](https://github.com/sourcenetwork/defradb/blob/develop/.github/workflows/check-vulnerabilities.yml)

3. **Dependency requirements**:  
   If a dependency requires a version bump and we cannot resolve the issue without upgrading, we will bump the version accordingly.

### Ideation / Proposal

The community follows the [Source Improvement Proposals](https://github.com/sourcenetwork/SIPs/) process for more comprehensive changes.

### Project management

- Use [Milestones](https://github.com/sourcenetwork/defradb/milestones) and a [project board](https://github.com/orgs/sourcenetwork/projects/3) to coordinate work on releases.

### Licensing

- Include the [BSL license header](./licenses/BSL.txt) at the top of every code file.
