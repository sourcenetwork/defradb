---
sidebar_label: Query Block
sidebar_position: 30
---
# Query Block

Query blocks are read-only GraphQL operations designed only to request information from the database, without the ability to mutate the database state. They contain multiple subqueries which are executed concurrently, unless there is some variable dependency between them.

Queries support database query operations such as filtering, sorting, grouping, skipping/limiting, aggregation, etc. These query operations can be used on different GraphQL object levels, mostly on fields that have some relation or embedding to other objects.