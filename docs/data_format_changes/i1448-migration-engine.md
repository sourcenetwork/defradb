# Add lens migration engine to defra

A new key-value was added to the datastore, it tracks the schema version of a datastore document and is required. If need be it could be set to the latest schema version for all documents, but that would prevent the migration of those records from their true version to that set version.