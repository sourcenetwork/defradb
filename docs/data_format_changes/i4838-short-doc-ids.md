# Short Doc IDs

Document storage keys now use a node-local short document ID instead of the public
document ID. The public document ID is derived from the first composite block CID
and is stored in a systemstore mapping.

Genesis document deltas also omit the public document ID so identical unsigned
creates produce identical genesis CIDs across nodes.
