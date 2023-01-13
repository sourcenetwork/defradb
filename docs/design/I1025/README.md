# Schema and Collection terminology cleanup

Currently there is a fair amount of duplication and confusion of terminology in Defra regarding the words schema and collection. We discussed this over zoom and decided that the definition would be as below:

> Schema: The global elements of a Collection description. A single Schema can be (in the future, currently unimplemented) used to define multiple Collections (both locally and globally).
>
> Collection: The local definition of a single complex database type set, analogous to a SQL table. The description of a collection includes the [global] Schema and any local-only elements such as indexes.

Parts of the codebase do not adhere to this definition and this should be corrected.

This document does not currently go into how they should be corrected, and only lists those that should.  How they should be corrected partly depends on the planned Schema Updates feature, and a work in progress design proposing a JSON based alternative to the current SDL format. Changes may also wish to bear in mind the planned allowance for multiple local Collections to be defined from a single Schema.

## client.AddSchema

`client.AddSchema` does not respect this definition. It accepts a `schemaString string` parameter, and uses it to define (local) Collections (the string can contain multiple type definitions). It then stores this local definition in the systemstore under the `/schema` prefix (discussed in [systemstore/schema](#systemstoreschema)).

It is a public entry point and the needs of external users should be taken into consideration when changing it.

## db.loadSchema

`db.loadSchema` loads an (local) Collection SDL from [systemstore/schema](#systemstoreschema) on database restart.

It is internal code called from one location and changing this should have little impact on anything.  That said, I believe it is currently quite poorly tested and particular care should be made with regards to testing any change here.

## systemstore/schema

A local Collection SDL is persisted in the systemstore under the `/schema` prefix. This is currently written to in [client.AddSchema](#clientaddschema) and read in [db.loadSchema](#dbloadschema) - called on database restart.

It is internal, but a breaking change, and will likely impact code within `client.AddSchema` and `db.loadSchema` when made. Otherwise the impact should be minimal.
