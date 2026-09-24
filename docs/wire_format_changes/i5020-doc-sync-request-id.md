# Add request ID to doc-sync requests

Adds a `RequestID` field (a random 16-byte nonce) to `docSyncRequest`. The
pubsub-rpc layer keys its response channel by a hash of the request bytes, so
concurrent syncs for the same docIDs used to collide and all but one would time
out. The nonce gives each request distinct bytes, mirroring the existing
`RequestID` on the KMS `fetchEncryptionKeyRequest`.

Old and new nodes remain compatible. The field is additive: an old node decoding
a new request ignores the unknown `requestID` key, and a new node decoding an
old request leaves `RequestID` empty. The reply type is unchanged. The fix takes
effect as soon as the requesting node is upgraded, since the collision happens
on the requester's side; responders never read the field.
