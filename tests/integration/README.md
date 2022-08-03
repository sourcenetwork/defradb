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
