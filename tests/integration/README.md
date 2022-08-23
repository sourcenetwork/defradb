## Database Integration Testing Guide

- We want to keep the mutation and query tests separate, here is what the folder
structure looks like currently:
```
    tests/integration
        ├── mutation/
        └── query/
```

- Every immediate directory under `tests/integration/mutation` and `tests/integration/query` should ONLY contain
    a single schema. For example:
`tests/integration/query/simple` and `tests/integration/query/complex` have different schemas.

- We can group different types of tests using the same schema into further sub-folders.
    For example:
    - `tests/integration/mutation/simple/create`: contains tests that 
        use the `simple` schema to test only the create mutation.
    - `tests/integration/mutation/simple/mix`: contains test that use
        the `simple` schema to test combination of mutations.

### Data Format Change Detection

Any test using the `ExecuteQueryTestCase` function in `tests/integration/utils.go` can also be used to assert that no undocumented breaking changes have been made in the active branch when compared to a target branch (default `develop`).

If the environment variable `DEFRA_DETECT_DATABASE_CHANGES` has been set, the test suite will run in this data format change detection mode instead of the standard test execution mode.  There is a CI build step that performs executes the tests in this mode for open pull requests.

When running a test in this mode, the following will happen:

1. Checkout and pull the latest version of the target branch into a temporary directory if it does not already exist.
2. Check for any new `.md` files in the `docs/data_format_changes` directory, if a new file is found - all tests will pass.
3. Create a new child process and execute the setup steps only (schema creation, database population, etc.) using the target branch code.
4. Execute the queries specified in the test using the current-branch/main-process against the database set up in step (3) and assert the results.

This should help reduce the risk of developers introducing undocumented changes to persisted data - something that could cause significant annoyance for users of defra, and loss of data.
