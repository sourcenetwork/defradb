# Namespace secondary index entries by epoch

Secondary index entries are now namespaced by an epoch, and each index tracks its live epoch in the
system store. This changes how index entries are keyed, so index data written by earlier versions is
not readable.
