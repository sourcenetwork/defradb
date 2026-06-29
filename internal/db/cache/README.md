# Cache

The `cache` package provides the means to cache anything in memory in Defra.

This document details how and why this *should* work, it does not necessarily detail how it is currently implemented.

## Details

The cache is split into several layers - `transaction`, `global` and `store`.  Each layer is free to perform additional cache layering should it chose, for example most on-disk `store` implementations will have some form of in-memory key-value caching.

## Transaction layer caching

The `transaction` cache layer sits on top of the `global` layer.  When it misses its cache, it will attempt to fetch the requested value from the `global` layer.

The `transaction` cache populates upon read from the `global` layer to ensure that transaction isolation is maintained, and that concurrent changes made to the `global` cache to not affect ongoing transactions.  The transaction cache needs to hold onto a version number that it can pass along to the `global` cache when reading, to ensure it only reads `global` values that existed when the transaction was created - preserving transaction isolation.

Each transaction-specific cache must be disposed of when the transaction is either discarded or committed.

Because the `transaction` cache only needs to hold on to a smaller subset of values held in the global cache, and is relatively short-lived, it is currently believed that it is not necessary to actively manage its size.

The `transaction` cache layer updates its values upon write to the transaction, there is little benefit in expiring them and requiring their re-fetching from lower layers, and this avoids us from having to worry about reading old, transaction-overwritten, values from the `global` and `store` layers.

## Global layer caching

The `global` cache layer sits on top of the `store` layer.  When it misses its cache, it will attempt to fetch the requested value from the `store` layer.

The `global` cache populates upon read from the `store` layer.  Each cached value is versioned, and will only be yielded to transactions with a higher version.

The `global` cache needs to have its size limited, this limit should be configurable.

## Store layer caching

The `store` layer sits at the bottom of the stack, and is responsible for fetching from the corekv store. The corekv store implementations may perform caching of their own.

Access to the underlying store is driven by the corekv transaction held on the context parameter.
