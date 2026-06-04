# Tests 

This directory contains two types of tests: benchmark tests (located in the bench directory) and integration tests (located in the integration directory). 
In addition to these, unit tests are also distributed among the other directories of the source code.

## Test Types

### Benchmark Tests

The bench directory contains benchmark tests that are used to measure and monitor the performance of the database.

### Integration Tests

The integration directory contains integration tests that ensure different components of the system work together correctly.

### Data Format Change Detection

Any test using the `ExecuteRequestTestCase` function in `tests/integration/utils.go` can also be used to assert that no undocumented breaking changes have been made in the active branch when compared to a target branch (default `develop`).

If the environment variable `DEFRA_CHANGE_DETECTOR_ENABLE` has been set, the test suite will run in this data format change detection mode instead of the standard test execution mode.  There is a CI build step that performs executes the tests in this mode for open pull requests.

When running a test in this mode, the following will happen:

1. Checkout and pull the latest version of the target branch into a temporary directory if it does not already exist.
2. Check for any new `.md` files in the `docs/data_format_changes` directory, if a new file is found - all tests will pass.
3. Create a new child process and execute the setup steps only (schema creation, database population, etc.) using the target branch code.
4. Execute the queries specified in the test using the current-branch/main-process against the database set up in step (3) and assert the results.

This should help reduce the risk of developers introducing undocumented changes to persisted data - something that could cause significant annoyance for users of defra, and loss of data.

### Unit Tests

Unit tests are spread throughout the source code and are located in the same directories as the code they are testing. 
These tests focus on small, isolated parts of the code to ensure each part is working as expected.

## Mocks

For unit tests, we sometimes use mocks. Mocks are automatically generated from Go interfaces using the mockery tool. 
This helps to isolate the code being tested and provide more focused and reliable tests.

To regenerate the mocks, run `make mocks`.

The mocks are typically generated into a separate mocks directory.

You can manually generate a mock for a specific interface using the following command:

```shell
mockery --name <interface_name> --with-expecter
```

Here, `--name` specifies the name of the interface for which to generate the mock.

The `--with-expecter` option adds a helper struct for each method, making the mock strongly typed.
This leads to more generated code, but it removes the need to pass strings around and increases type safety.

For more information on mockery, please refer to the [official repository](https://github.com/vektra/mockery).

## License

The contents of this directory form the DefraDB integration test suite.

Unlike the rest of the repository, which is licensed under the
Business Source License (BSL), the test suite is dual-licensed:

- GNU Affero General Public License v3 (AGPLv3)
- Business Source License 1.1 with additional use restrictions

See LICENSE in this directory for details.