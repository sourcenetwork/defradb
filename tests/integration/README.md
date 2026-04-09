# Database Integration Testing Guide

This guide documents the high level concepts used by the core Defra test suite.

The same concepts discussed here are shared with other database projects, such as Lens and corekv, as well as some client specific tests that can be found in `./cli/test/`

## Concepts

Integration testing in Defra is driven by two key concepts - `actor-action based testing` and `complexity multipliers`.  The combined end goals of the two is to minimize the amount of tests that need to be written, ensure that tests are decoupled from the internal implementation, and are easily read, including by newcomers to the codebase.

### Actor-action based testing

Tests are declared as a set of actions performed by a set of external actors.  For example a test might specify that identity 1 creates a collection, then adds a document.

Each action is a reusable data struct, that matches as closely as practical to how an actor might interact with the database.

Actions are declared ahead of time, and only executed after all actions have been declared and modified by `complexity multipliers`.

Many of the actions have been migrated to `./tests/action`, this allows the actions to also be used to write benchmark tests.

### Complexity multipliers

The Defra integration test suite evolved organically into something that conceptually matched [testo](https://github.com/sourcenetwork/testo/) complexity multipliers, but the core suite has not yet been migrated to the testo system.  The CLI integration tests, Lens, and corekv tests do make direct use of testo.

Testo defines a complexity multiplier as:
> A complexity multiplier represents a concept that multiplies the surface area and complexity
> of other proximal features, for example, database transactions are complexity modifiers, as
> when adding many new database actions such as a new filter operation, the new action must be
> tested both with, and without a transaction - the transaction concept multiplies the complexity
> of the system.

Defra has designed it's test suite around this principle, allowing developers to write a single, relatively simple test, that will automatically provide coverage for a wide range of scenarios.  For example, a test may declare a single `AddDocument` action, the complexity multiplier system allows that single test to cover a matrix of possibilities - adding via the embedded Go client, the HTTP and CLI clients, adding via the collection and GQL APIs, encryption, document signing, different storage engines, etc, etc - all via the same, simple, test declaration.

The test suite executes a test once, using a single configuration, when `go test` is called.  The configuration is selected based on a set of potential environment variables and their defaults.  Unlike testo, each configuration option has its own variable, and as these are somewhat liable to change, perhaps the best way of discovering them is by looking at the `env:` section in `./.github/workflows/test-coverage.yml`.  Once you have found the name of the variable you are interested in, you can search for it's usage within the `./tests` directory, some of them are documented.

Some of these have helper declarations in the make file.

For example, to execute tests via the http client, you can run:
```bash
DEFRA_CLIENT_HTTP=true go test ./...
```
or
```bash
make test:http
```

Multiple multipliers can be combined, for example, the below will test the http client via GQL mutations, on a badger-file store:
```bash
DEFRA_CLIENT_HTTP=true DEFRA_MUTATION_TYPE=gql DEFRA_BADGER_FILE=true go test ./...
```

## Directory structure

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
