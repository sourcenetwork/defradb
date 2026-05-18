# PatchCollection fails to set CRDT if secondary relationship exists on collection

This is not a true breaking change, but in previous Defra versions patching new fields onto a collection that already had a secondary relation would not be provided a default CRDT, and thus would not be writeable.

Those new fields cannot be recovered and should be deprecated.
