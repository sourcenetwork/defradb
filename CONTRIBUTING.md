## Welcome
You have made your way to the contributing guidelines for DefraDB! You are about to be part of a wave making databases decentralized and more powerful.

This document is to help you with the process of contributing to the project. First recommendation is to become familiar with the [project documentation](https://docs.source.network/). Familiarity with [Git](https://training.github.com/), [Go](https://go.dev/doc/install), and [Github](https://docs.github.com/) is assumed.

All contributions are appreciated, whether it's identifying problems, highlighting missing features, or contributing to the codebase in straightforward or complex ways.

You're invited to join the [Source Network Discord](discord.source.network), where you can discuss ideas, ask questions, and find inspiration for what to build next.


## Getting started
Obtain the repository, then build and run it.
```shell
git clone https://github.com/sourcenetwork/defradb.git
cd defradb
make start
```

Refer to the `README.md` and project documentation for various usage examples.


## Development flow

Methodologies the project follows:

- *Issue-driven development*: Every pull request links to issue(s).
- [Squash and merge](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/incorporating-changes-from-a-pull-request/about-pull-request-merges): Commits of a pull request are squashed into one before being merged onto the `develop` branch.
- [Conventional commits](https://www.conventionalcommits.org/en/v1.0.0/): Every commit message is of the `<type>: <description>` format in which type is one of `feat`, `fix`, `tools`, `docs`, `perf`, `refactor`, `test`, `ci`, `chore`.

Basic development flow:
1. Make changes
2. If applicable, write tests of the new/changed behavior
3. Ensure that `make test` and `make lint` are passing

Creating an issue:
1. On https://github.com/sourcenetwork/defradb/issues, click "New issue".
2. Select the relevant issue type.
3. Fill the issue template.

Submitting a contribution:

1. Fork the repository and with your changes create a branch following the convention `<your-name>/<type>/<description>`, for example `octavio/feat/compression-lru-data`.
2. Create a pull request targeting the `develop` branch. Link the relevant existing issue(s), and create one if no related issue exists yet. Follow the instructions of the pull request template. Use a verb in the PR title to describe the code changes.
3. Read through and accept the Contributor License Agreement, if it's your first contribution.
4. Request review from the *database-team*, and discuss and adapt the pull request accordingly.
5. If approved, click 'Squash and merge' to squash it into one commit and to be merged in `develop`. Make sure the title of the commit follows the conventional commits convention.


## Testing

`make test` to run the unit tests and integration tests.

`make lint` to run the linters.

`make bench` to run the benchmark suite. To assess the difference between a branch's results and the `develop` branch's results, execute the suite on both outputting the results to files, and compare the two with e.g. `benchstat develop.txt current.txt`. To install `benchstat` use `make deps:bench`.

`make test:changes` to run a test suite detecting breaking changes. Breaking changes need to be accompanied by documentation in `docs/data_format_changes/` for the test to pass.


## Documentation
- The overall project's documentation is found at [docs.source.network](https://docs.source.network), and its source at [github.com/sourcenetwork/docs.source.network](https://github.com/sourcenetwork/docs.source.network).
- The code documentation, doc comments, can be viewed as a website:
	```shell
	go install golang.org/x/pkgsite/cmd/pkgsite@latest
	cd your-path-to/defradb/
	pkgsite
	# open http://localhost:8080/github.com/sourcenetwork/defradb
	```
	- [go.dev/doc/comment](https://go.dev/doc/comment) has guidelines on writing Go doc comments.
- `docs/cmd/` is where auto-generated documentation of the `defradb` command-line program is.
- `docs/data_format_changes/` details the historical breaking changes to data persistence.


## Additional information

This section includes good-to-know information for advanced contributors.

Release process and versioning:
- The project follows [Semantic Versioning](https://semver.org/).
- `CHANGELOG.md` is automatically generated at the end of the release process, using `make chglog`, and finally manually verifying its content. This is possible because of conformance to *conventional commits*.

Dependency management:
- `go mod tidy` should be performed by pull requests changing dependency requirements.
- Using `dependabot` for the automatic creation of pull requests updating dependencies.

Peer review of pull requests:
- Using the [Conventional Comments](https://conventionalcomments.org/) methodology is recommended.

Licensing:
- A license header must be included at the top of every code file.

Project management:
- [Milestones](https://github.com/sourcenetwork/defradb/milestones)  and a [project board](https://github.com/orgs/sourcenetwork/projects/3) are used to coordinate work on releases.