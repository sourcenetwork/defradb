# Remove the first CRDT byte from field encoded values

The first CRDT byte was legacy code and no longer necessary as we have this information independently available via the client.FieldDescription, since the FieldDescription.Typ is the exact same value.