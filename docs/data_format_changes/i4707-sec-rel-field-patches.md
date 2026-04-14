# PatchCollection fails to set CRDT if secondary relationship exists on collection

New fields added to collections that already had a secondary relation would fail to set a default CRDT if one was not explicitly provided.  Those newly added fields will not work properly and should be deprecated.
