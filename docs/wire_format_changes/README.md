# Wire format changes

Any change to a type a node sends to another node (its fields, their names, or
their order) can break communication with nodes running an older version. When
the wire snapshot changes, add a short markdown file here describing the change
and its cross-version impact: can old and new nodes still talk, and if not, what
is the migration or version bump.

The wire types are those registered in `internal/wire` (see `wire.Register`). The
snapshot lives at `internal/wire/snapshottest/wire_snapshot.golden`. After an
intentional change, regenerate it with:

    make test:wire-snapshot-update

and commit the updated golden together with your note here.
