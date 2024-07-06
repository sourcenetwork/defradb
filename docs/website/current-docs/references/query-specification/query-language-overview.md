---
sidebar_label: Query Language Overview
sidebar_position: 10
---
# Query Language Overview

The DefraDB query language (DQL) is a GraphQL defined API which is used to access and query data residing inside a DefraDB node.

[GraphQL](https://graphql.org) is an open-source query language for APIs, built for making APIs fast, flexible, and developer friendly. Databases such as [DGraph](https://dgraph.io/) and [Fauna](https://fauna.com) use GraphQL API as a query language to read and write data to/from the database. 
- DGraph is a distributed, high throughput graph database. 
- Fauna is a transactional database delivered as a secure, web-native API GraphQL.

DefraDB (while using GraphQL) is designed as a document storage database, unlike DGraph and Fauna. DQL exposes every functionality of the database directly, without the need for any additional APIs. The functionalities include:
- Reading, writing, and modifying data.
- Describing data structures, schemas, and architecting data models (via index's and other schema independent, application-specific requirements).
  
**Exception**: DefraDBs PeerAPI is used to interact with other databases and with the underlying CRDTs (for collaborative text editing).

Our initial design relies only on the currently available GraphQL specification (version tagged June 2018 edition). Initially, the GraphQL Query Language will utilize standard GraphQL Schemas, with any additional directives exposed by DefraDB. DefraDBs CRDT types will initially be automatically mapped to GraphQL types.