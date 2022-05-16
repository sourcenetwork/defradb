## Testing Guide

- We want to keep the mutation and query tests separate, here is what the folder
structure looks like currently:
```
    testing/db
        ├── mutation/
        └── query/
```

- Every immediate directory under `testing/db/mutation` and `testing/db/query` should ONLY contain
    a single schema. For example:
`testing/db/query/simple` and `testing/db/query/complex` have different schemas.

- We can group different types of tests using the same schema into further sub-folders.
    For example:
    - `testing/db/mutation/simple/create`: contains tests that 
        use the `simple` schema to test only the create mutation.
    - `testing/db/mutation/simple/mix`: contains test that use
        the `simple` schema to test combination of mutations.
