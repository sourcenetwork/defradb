## Contributing to DefraDB

Thank you for your interest in contributing to DefraDB! You're about to join a wave of innovation in decentralized and powerful databases.
This guide will help you navigate the contribution process, whether you're identifying issues, suggesting features, or contributing code.


It is recommended that before you begin, familiarize yourself with the following:
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

- The `docs/website/references/http/openapi.json/` directory contains auto-generated http api documentation.

- The `docs/website/references/cli/` directory contains auto-generated cli documentation.

### Manpages

We do support manpages, in order to get them working please do:

```sh
make docs:manpages
```

Then the man pages will be in `build/man/` directory.

You can install them manually on your system, or if you are on linux you can do:

```sh
make install:manpages
```

## Getting Started

To get started, clone the repository, build, and run it:
```shell
git clone https://github.com/sourcenetwork/defradb.git
cd defradb
make start
```

Refer to the [`README.md`](./README.md) and project documentation for usage examples.

## Creating an Issue

1. Go to [github.com/sourcenetwork/defradb/issues](https://github.com/sourcenetwork/defradb/issues), and click "New issue".
2. Select the relevant issue type.
3. Fill out the issue template.

## Git Workflow

### Fork Flow vs Branch Flow

We have adopted the Git Fork Flow as our primary development workflow for both internal and external developers. This decision was made to ensure a consistent and streamlined approach to contributions across our project.

### Fork Flow for All Contributors

All developers, whether internal or external, are expected to use the Fork Flow for their contributions. This involves:

1. Forking the main repository.
2. Cloning your forked repository locally.
3. Creating a feature branch on the clone of your forked repository.
4. Making changes on the feature branch.
5. Writing tests for the changed behavior, if applicable.
6. Ensuring that all the `make` target checks are passing:
    - `make docs`
    - `make mocks`
    - `make lint`
    - `make tidy`
    - `make test`
    - `make test:changes`
7. Committing your changes on to the feature branch.
8. Pushing your feature branch with the commit changes, on to your fork.
9. [Opening a pull request](#opening-pull-request) from your branch on your fork, targeting the `develop` branch of the main repository.

### Limited Use of Branch Flow

In certain circumstances, internal developers may need to use the Branch Flow, particularly for CI-related updates that require direct access to the main repository. If an internal developer opts to use the Branch Flow, it will look something like this:

1. Clone the main repository locally.
2. Create a feature branch on the clone, adhering to the following branch naming convention
    ```
     <dev-name>/<label>/<description>
    ```
    For example: `lone/ci/update-test-workflow-action`

3. Make changes on the feature branch.
4. Write tests for the changed behavior, if applicable.
5. Ensure that all the `make` target checks pass:
    - `make docs`
    - `make mocks`
    - `make lint`
    - `make tidy`
    - `make test`
6. Commit your changes on to the feature branch.
7. Push your feature branch with the commit changes, on to the main repository.
8. [Open a pull request](#opening-pull-request) on the main repository targeting the `develop` branch.

**Note:** The use of Branch Flow should be limited even amongst internal developers.

## Opening Pull Request

### Link with relevant issue(s)
Since we follow **Issue-driven development**, every pull request **must** be [linked](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/using-keywords-in-issues-and-pull-requests#linking-a-pull-request-to-an-issue) to one or more issues.
An issue can be linked by adding the issue number after using any of the [resolving keywords](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/using-keywords-in-issues-and-pull-requests#linking-a-pull-request-to-an-issue):
- close
- closes
- closed
- fix
- fixes
- fixed
- resolve
- resolves
- resolved

For example, this PR links to issue `#123`:

```
## Relevant issue(s)

Resolves #123
```

Note: There is a `Relevant issue(s)` section at the top of the PR description template just for this purpose.

### Title Format

The pull request title **must** use a leading label from a fixed set of labels (most labels are inspired from **[conventional commits](https://www.conventionalcommits.org/en/v1.0.0/)** style) followed by the description.

The pull request title must be in this format `<label>: <Description>` where:

- `<label>` is one of:
  - `chore`: A routine or maintenance type of task.
  - `ci`:  A CI/CD related task.
  - `docs`: A documentation relation task.
  - `feat`: Task adding a new feature.
  - `fix`: Task related to a fix that was done.
  - `perf`: Performance related task.
  - `refactor`: A refactor related task.
  - `test`: Task that involves testing related work.
  - `tools`:  Task with tooling-related work.
  - `bot`: Automated task (shouldn't be used manually by non-bot authors).
- Right after `<label>` there is a `:` and then a single space (` `).
  - Invalid example with no space: `fix:invalid example`
- After the single space we have `<Description>`
- First letter of the `<Description>` **must** be a capital letter.
  - Invalid example with lowercase: `fix: invalid example`
- First word of the `<Description>` **should** be an action verb.
  - Interesting [example list of verbs / first word to use](https://gist.github.com/scmx/411f6fea4ee3832806720d536a7d5d8f)
- Last character **must** be an alphanumeric character (`a-z` | `A-Z` | ` 0-9`).
- If label is not `bot`, the entire title **must** not be greater than 60 characters.

Note: More examples of valid and invalid PR titles can be found in [tools/scripts/scripts_test.sh](https://github.com/sourcenetwork/defradb/blob/develop/tools/scripts/scripts_test.sh).

### Sign the CLA
Read and accept the Contributor License Agreement (for first-time contributors).

## Managing Pull Request

### Asking for review
- Request a review from the *database-team*.
- Discuss and adapt the pull request as needed by following the commenting etiquette defined below.

### Commenting Etiquette

It can sometimes be confusing to judge if a reviewer's comment is a blocking, or a non-blocking comment. Therefore we have been inspired by **[conventional comments](https://conventionalcomments.org/)** to adapt the following labels:

Label | Meaning | Action
--- | --- | ---
`thought` | These comments are a dump of thoughts. Which may or may not be within the scope of this PR. These can be to provide context or just to make brain juices flow. |  No action is required.
`question` | These comments are questions that the reviewer has. They may later evolve into other types of comments once clarity is achieved. | Answer the question, or guide the reviewer to a resource that answers the question.
`nitpick` | Minor, nitpick suggestion that the reviewer might have. | Can be ignored or accepted, without any followups.
`suggestion` | Non-blocking suggestions that the reviewer might have. | Either accept the suggestion, Or provide context on why this shouldn't be done.
`todo` | Blocking suggestions, that must be resolved before merge. | Must resolve these before the PR can be merged. If resolving in another PR, make an issue and link it.
`issue` | Blocking major issue, that **MUST** be resolved in this PR. | Must resolve in this PR.


## Merging Pull Request

### Pass CI Checks

These are the required CI checks, that **MUST** pass in order to have the PR merged.

Check Type | Description | Resolve
--- | --- | ---
`Check Data Format Changes` | Ensures no data format changes occurred. If they did occur, resolve by documenting the changes in this directory: `docs/data_format_changes/`, more instructions [here](./docs/data_format_changes/README.md).  | To run the change detector, simple do `make test:changes`. More about how it works [here](./tests/change_detector/README.md).
`Check Documentation` | Ensures all documentation is up to date. Some documentation might need to be re-generated depending on the changes.  | To generate the documentation, do `make docs`.
`Check Lint` | Ensures all linting rules are adhered to. | To install the linting tools, do `make deps:lint`. To see the lint failures do `make lint`. In some cases to auto fix linter failures do `make lint:fix`.
`Check Mocks` | Ensures all mocks are up to date. Some mocks might need to be regenerated depending on the changes.  | To generate the mocks, do `make mocks`.
`Check Tidy` | Ensures `go mod tidy` is in a clean state. | Run `make tidy`.
`Start Binary` | Ensures the binary actually can be started after building. | To build, do `make build` then start by `./build/defradb start --no-keyring`.
`Test Coverage` | Ensures the different combinations of tests work (also generates coverage). | To run the tests, do `make test`. However you may need to use additional `ENV` variables to trigger specific kinds of tests. For more details on what each test job tests, please look at the [workflow file](.github/workflows/test-coverage.yml).
`Validate Container` | Ensures the container file did not break. | This can be resolved by ensuring the [container file](./tools/defradb.containerfile) is still valid.
`Validate Title` | Ensures the PR title adheres to the rules described [above](#title-format). | This is ran automatically when title changes. See how it works [here](./tools/scripts/validate-conventional-style.sh).

### Ready to merge?

A PR is ready to merge when:
- There is at least 1 approval.
- There is no `DO NOT MERGE` label.
- It is rebased with the upstream `develop` branch.
- All required checks are passing

Note: We follow the **[Squash and Merge](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/incorporating-changes-from-a-pull-request/about-pull-request-merges)** strategy.

#### _Internal developer(s)_:
Click the "Squash and merge" to combine the commits and merge into the `develop` branch.

#### _External contributor(s)_:
Inform a maintainer that you are ready to have to PR merged, and the maintainer will merge the PR for you.

## Extras

### GoLang Version Bumping Policy
Currently on DefraDB our version bumping policy is:

1) Strictly one version behind (-1) of the latest go version. The lower limit of this was established because a Go release becomes unsupported the moment second new major release is cut. The upper limit was due to us not wanting to be on the latest release to avoid bleeding edge instabilities that weren't caught.

2) If there are any vulnerabilities that are fixed in the latest version shown by `govulncheck`, which have not made it to the patch of the current go version we are on (usually under 24hrs of the vulnerability being known the patch for the current version is available in which case we don't bump). 
Link to vulnerability check action: https://github.com/sourcenetwork/defradb/blob/develop/.github/workflows/check-vulnerabilities.yml

3) If a dependency has a strict need to bump, and DefraDB can't resolve it without bumping.

### Ideation / Proposal
The community follows the [Source Improvement Proposals](https://github.com/sourcenetwork/SIPs/) process for more comprehensive changes.


### Project management:
- Use [Milestones](https://github.com/sourcenetwork/defradb/milestones) and a [project board](https://github.com/orgs/sourcenetwork/projects/3) to coordinate work on releases.

### Licensing:
- Include the [BSL license header](./licenses/BSL.txt) at the top of every code file.
