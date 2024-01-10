# Change CRDT encoded data struct fields

The composite CRDT delta struct no longer hosts the changed properties of the document. The removes the leakage of field level values for when we implement field level access control.
