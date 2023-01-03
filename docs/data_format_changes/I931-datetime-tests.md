# New DateTime Field

Addition of the new DateTime field changes the test schema of the `integration/query/simple` tests which changes CIDs because `SchemaID` is serialized into the delta payload. The new schema (and `SchemaID`) triggers the change-detection script.