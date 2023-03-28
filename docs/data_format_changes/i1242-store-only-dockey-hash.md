# Store only dockey hash to delta (block) storage

Since data that we store is supposed to be synchronized between nodes, we can
store there only global identifies. Things like CollectionID, FieldID, etc are local
and as such can't be stored. On the other hand, dockey hash is global and that's
why it's the only thing that makes sense to store.
