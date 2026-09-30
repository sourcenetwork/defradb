# Short Doc IDs

Document storage keys now use a node-local short document ID instead of the public
document ID. The public document ID is derived from the first composite block CID
and is stored in a systemstore mapping.

Document deltas no longer encode the public document ID. For a genesis composite this
keeps creates deterministic across nodes: when the content is neither signed nor
encrypted, two nodes produce the same genesis CID and therefore the same public document
ID. Signatures and encryption are embedded in the composite block, so signed or encrypted
documents produce a different public document ID per signing identity. Signing is enabled
by default, so this cross-node determinism applies only to explicitly unsigned, unencrypted
documents.
