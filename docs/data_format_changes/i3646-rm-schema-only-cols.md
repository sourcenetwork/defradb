# Replace schema-only collections with boolean

The way Collection/SchemaDescriptions are serialized and saved to disk has changed.  Embedded types (e.g. inner View objects) are no longer represented only by a schema, they now have a collection description with a IsEmbeddedOnly boolean set to true.
