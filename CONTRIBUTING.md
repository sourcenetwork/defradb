## Welcome
You have made your way to the contributing guidelines of DefraDB! You are about to be part of a wave making databases decentralized and more powerful.

This document will guide you through the process of contributing to the project. It is recommended that you familiarize yourself with the [project documentation](https://docs.source.network/), [Git](https://training.github.com/), [Go](https://go.dev/doc/install), and [Github](https://docs.github.com/) before getting started.

All contributions are appreciated, whether it's identifying problems, highlighting missing features, or contributing to the codebase in simple or complex ways.

You are encouraged to join the [Source Network Discord](discord.gg/w7jYQVJ) to discuss ideas, ask questions, and find inspiration for future developments.

## Getting started
To get started, clone the repository, build, and run it:
```shell
git clone https://github.com/sourcenetwork/defradb.git
cd defradb
make start
```

Refer to the [`README.md`](./README.md) and project documentation for usage examples.

## Development flow

The project follows these methodologies:

- **Issue-driven development**: Every pull request is linked to one or more issues.
- **[Squash and merge](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/incorporating-changes-from-a-pull-request/about-pull-request-merges)**: Commits of a pull request are squashed into one before being merged onto the `develop` branch.
- **[Conventional commits](https://www.conventionalcommits.org/en/v1.0.0/)**: Every commit message is in the `<type>: <description>` format, where "type" is one of `feat`, `fix`, `tools`, `docs`, `refactor`, `test`, `ci`, `chore`, `bot`.


To create an issue:
1. Go to [github.com/sourcenetwork/defradb/issues](https://github.com/sourcenetwork/defradb/issues), and click "New issue".
2. Select the relevant issue type.
3. Fill out the issue template.

Follow this basic development flow:
1. Make changes.
2. Write tests for the changed behavior, if applicable.
3. Ensure that `make test` and `make lint` are passing.

To submit a contribution:
1. Create a branch with your changes following the `<your-name>/<type>/<description>` convention, e.g., `octavio/feat/compression-lru-data`.
2. Create a pull request targeting the `develop` branch. Link to the relevant existing issue(s), and create one if none exists. Follow the pull request template instructions and use a verb in the PR title to describe the code changes.
3. Read and accept the Contributor License Agreement, if it's your first contribution.
4. Request a review from the *database-team*. Discuss and adapt the pull request as needed.
5. Once approved, click "Squash and merge" to squash the commits into one and merge them into the `develop` branch. Ensure that the commit title follows the *Conventional Commits* convention.


When introducing breaking changes, include the `BREAKING CHANGE` keyword in the commit message body, followed by a description of the change. This helps keep track of changes that may require additional attention or migration steps.

## Testing

Run the following commands for testing:

- `make test` to run unit and integration tests.
- `make lint` to run linters.
- `make bench` to run the benchmark suite. To compare a branch's results with the `develop` branch results, execute the suite on both branches, output the results to files, and compare them with a tool like `benchstat` (e.g., `benchstat develop.txt current.txt`). To install `benchstat`, use `make deps:bench`.
- `make test:changes` to run a test suite detecting breaking changes. Accompany breaking changes with documentation in `docs/data_format_changes/` for the test to pass.

### Test prerequisites

The following tools are required in order to build and run the tests within this repository:

- [Go](https://go.dev/doc/install)
- Cargo/rustc, typically installed via [rustup](https://www.rust-lang.org/tools/install)

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

The `docs/cmd/` directory contains auto-generated documentation for the `defradb` command-line program.

The `docs/data_format_changes/` directory provides details on the historical breaking changes to data persistence.

## Additional information

This section contains useful information for advanced contributors.

Dependency management:
- Run `go mod tidy` for pull requests that change dependency requirements.
- The project uses `dependabot` to automatically create pull requests for updating dependencies.

Peer review of pull requests:
- It is recommended to use the [Conventional Comments](https://conventionalcomments.org/) methodology.

Licensing:
- Include the [BSL license header](./licenses/BSL.txt) at the top of every code file.

Project management:
- Use [Milestones](https://github.com/sourcenetwork/defradb/milestones) and a [project board](https://github.com/orgs/sourcenetwork/projects/3) to coordinate work on releases.

The community follows the [Source Improvement Proposals](https://github.com/sourcenetwork/SIPs/) process for more comprehensive changes.