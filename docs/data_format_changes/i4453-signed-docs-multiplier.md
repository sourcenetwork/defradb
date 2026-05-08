# Signed-docs test multiplier

Adds a `signed-docs` test multiplier that flips `TestCase.EnableSigning = true`
across the integration suite. Signing attaches a `Signature` link to composite
blocks and priority-1 field blocks in the Merkle DAG, which changes block bytes
and therefore CIDs. This is intentionally incompatible at the block level with
databases created without signing, so the change detector cannot verify
compatibility between the signing-enabled and signing-disabled modes.
