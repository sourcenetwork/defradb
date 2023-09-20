# Change Detector

The change detector is used to detect data format changes between versions of DefraDB.

## How it works

The tests run using a `source` and `target` branch of DefraDB. Each branch is cloned into a temporary directory and dependencies are installed.

The test runner executes all of the common test packages available in the `source` and `target` tests directory.

For each test package execution the following steps occur:

- Create a temporary data directory. This is used to share data between `source` and `target`.
- Run the `source` version in setup only mode. This creates test fixtures in the shared data directory.
- Run the `target` version in change detector mode. This skips the setup and executes the tests using the shared data directory.
