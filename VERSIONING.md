# Versioning

This document outlines our versioning policy.

Generally, we will be incrementing the Defra release version in strict accordance with [semver](https://semver.org/):

> Given a version number MAJOR.MINOR.PATCH, increment the:
>
> 1. MAJOR version when you make incompatible API changes
> 2. MINOR version when you add functionality in a backward compatible manner
> 3. PATCH version when you make backward compatible bug fixes

However, in some areas the line is blurry and controversial, and very occasionally deemed to be impractical in the short-term.  The rest of this document outlines these special areas and our versioning policy on them.

## Errors

At the moment, our errors are largely string based - they are driven by a single concrete implementation of the standard Go `error` interface [here](./errors/defraError.go#L60).

The vast majority of our errors are declared by an `errors.go` file inside the relevant package, for example in [client](./client/errors.go).  The exact strings are private, but the error instance, and function returning those instances are public to the package.

We as a team have decided that, whilst the exact strings are user-visible via the clients, we are not going to protect them with semver.  If we feel the need, for example, to fix a typo, we may change these strings as part of a `PATCH` or `MINOR` version.

We will not make significant changes to the structure of errors returned by any of our functions.  For example, if a function returns an `client.ErrValueTypeMismatch` wrapped by an `client.ErrInvalidJSONPayload` - removing or replacing either one of these will only be done as part of a `MAJOR` version increment.

## The C embedded client

We made a mistake when designing many of the function signatures that form the [C client](./cbindings).  Many of the functions have parameters that do not pair up with their Go equivalents - they take individual formal parameters, where the Go function takes an [options](./client/options) struct.

This creates significant friction for us when introducing new function options, as adding a new parameter to a semver-protected function requires a `MAJOR` version increment.

At the time of writing, we do not have any known external users of the C bindings, and have chosen to temporarily omit the C bindings from our semver protection. We expect this problem to be resolved early within the development of Defra v1.2, and expect to protect the C bindings with our usual semver guarantees from v1.2 onwards.
