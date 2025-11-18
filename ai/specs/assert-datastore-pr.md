# Add Datastore test action for direct datastore value assertions

This PR introduces a new Datastore test action to the DefraDB integration test framework that enables direct fetching and assertion of values from the datastore.
The action provides a flexible builder pattern for constructing keys.

The implementation adds a Datastore struct that accepts a KeyBuilder interface for key construction and uses gomega.OmegaMatcher for value assertions.
This allows tests to verify the presence or absence of specific keys in the datastore and assert their values with flexible matchers.

A builder pattern API is provided through the NewKey() factory function, which supports building different key types:

- DatastoreDoc() builder creates keys for document field values
- DatastoreIndex() builds keys for index entries

Both builders use collection and document indices that are resolved to actual IDs at runtime, maintaining consistency with the existing test framework patterns.
