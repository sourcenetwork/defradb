# Remove collection name from schema ID generation

The collection name was removed from the schema ID generation, this caused test schema IDs and commit CIDs to change.  Will also impact production systems, as identical schemas created on different defra versions would not have the same IDs.