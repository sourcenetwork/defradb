# Data Definition in a DefraDB instance

Data held in a DefraDB instance is organized into [collections](#collections) of documents.  [Collections](#collections) are [local](#local-definitions) groupings of documents that share the same [globally](#global-definitions) defined shape declared by a [schema](#schemas).

## Local definitions

Local definitions are specific to the node you are directly working with, they are not shared with, or assumed to be the same on other nodes in the network.

Splitting local elements out from the global ones allows some local customization to the way data is organized within any given node.  It also minimizes the amount of 'stuff' that must be kept consistent across the decentralized network in order to have a well behaving database.

Local data definitions are always defined on the [collection](#collections).

Examples include indexes, field IDs, and [lens transforms](https://docs.source.network/defradb/guides/schema-migration).

## Global definitions

Global definitions are consistent across all nodes in the decentralized network. This is enforced by the use of things like CIDs for schema versions.  If a global definition was to differ across nodes, the different variations will be treated as a completely different definitions.

Global data definitions are always defined on the [schema](#schemas).

Examples include field names, field kinds and [CRDTs](https://docs.source.network/defradb/guides/merkle-crdt).

## Collections

Collections represent [local](#local-definitions), independently queryable datasets sharing the same shape.

Collections are defined by the `CollectionDescription` struct.  This can be mutated via the `PatchCollection` function.

A collection will always have a [global](#global-definitions) shape defined by a single [schema](#schemas) version.

### Versions

`CollectionDescription` instances may be active or inactive.  Inactive `CollectionDescription`s will not have a name, and cannot be queried.

When a new [schema](#schemas) version is created for a schema that has a collection defined for it, a new `CollectionDescription` instance will be created and linked to the new schema version.  The new `CollectionDescription` instance will share the same root ID as the previous, and may be active or inactive depending on what arguments the user defining the new schema specified.

[Lens migrations](https://docs.source.network/defradb/guides/schema-migration) between collection versions may be defined.  These are, like everything on the collection, [local](#local-definitions).  They allow transformation of data between versions, allowing documents synced across the node network at one schema version to be presented to users at **query time** at another version.

### Collection fields

The set of fields on a `CollectionDescription` defines [local](#local-definitions) aspects to [globally](#global-definitions) defined fields on the collection's [schema](#schemas).  The set may also include local-only fields that are not defined on the schema, and will not be synced to other nodes - currently these are limited the secondary side of a relationship defined between two collections.

### Views

Collections are not limited to representing writeable data.  Collections can also represent views of written data.

Views are collections with a `QuerySource` source in the `Sources` set.  On query they will fetch data from the query defined on `QuerySource`, and then (optionally) apply a [Lens](https://github.com/lens-vm/lens) transform before yielding the results to the user.  The query may point to another view, allowing views of views of views.

Views may be defined using the `AddView` function.

### Embedded types

Some fields on a collection may represent a complex object, typically these will be a relationship to another collection, however they may instead represent and embedded type.

Embedded types cannot exist or be queried outside of the context of their host collection, and thus are only defined as a [global](#global-definitions) shape represented by a [schema](#schemas) only.

Related objects defined in a [view](#views) are embedded objects.

## Schemas

Schemas represent [global](#global-definitions) data shapes.  They cannot host document data themselves or be queried, that is done via [collections](#collections).

Schemas are defined by the `SchemaDescription` struct.  They are immutable, however new versions can be created using the `PatchSchema` function.

Multiple [collections](#collections) may reference the same schema.
