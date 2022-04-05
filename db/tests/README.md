## Testing Guide

- We want to keep the mutation and query tests seperate, here is what the folder
structure looks like currently:
```
    db/tests
        ├── mutation/
        └── query/
```

- Every immediate directory under `db/tests/mutation` and `db/tests/query` should ONLY contain
    a single schema. For example:
`db/tests/query/simple` and `db/tests/query/complex` have different schemas.

- We can group different types of tests using the same schema into furthur sub-folders.
    For example:
    - `db/tests/mutation/simple/create`: contains tests that 
        use the `simple` schema to test only the create mutation.
    - `db/tests/mutation/simple/mix`: contains test that use
        the `simple` schema to test combination of mutations.
