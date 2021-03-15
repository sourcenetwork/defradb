<p align="center"> 
<img height="120px" src="docs/DefraDB_Full-v2-cropped.png">
</p>

#
The DefraDB is a Peer-to-Peer Edge Database, with the interface of a NoSQL Document Store. DefraDB's data model is backed by [MerkleCRDTs](https://arxiv.org/pdf/2004.00107.pdf) for a multi write-master architecture. It is the core data storage system for the [Source](https://source.network/) Ecosystem, built with [IPFS](https://ipfs.io/) technologies like [IPLD](https://docs.ipld.io/) and [LibP2P](https://libp2p.io/) and featuring Semantic web3 properties. You can read the [Technical Overview here](https://docsend.com/view/mczj7ic4i3kqpq7s).

Table of Contents
=================

* [Early Access](#early-access)
* [Installation](#installation)
	* [Compile](#compile)
* [Getting Started](#getting-started)
	* [Add a Schema type](#add-a-schema-type)
	* [Create a Document Instance](#create-a-document-instance)
	* [Query our documents](#query-our-documents)
	* [Interact with Document Commits](#interact-with-document-commits)
* [Query Documenation](#query-documenation)
* [CLI Documentation](#cli-documentation)
* [Next Steps](#next-steps)
* [Licensing](#licensing)
* [Contributors](#contributors)
* [Futher Reading](#futher-reading)
	* [Technical Specification Doc](#technical-specification-doc)
	* [Design Doc](#design-doc)

## Early Access
DefraDB is currently in a *Early Access Alpha* program, and is not yet ready for production deployments. Please reach out to the team at [Source](https://source.network/) by emailing [hello@source.network](mailto:hello@source.network) for support with your use-case and deployment.

## Installation
To install a DefraDB node, you can download the pre-compiled binaries available on the releases page, or you can compile it youself if you have a local [Go Toolchain](https://golang.org/) installed.

### Compile
```
go install github.com/sourcenetwork/defradb/cli/defradb
```
or
```
git clone git@github.com:sourcenetwork/defradb-early-access.git
mv defradb-early-access defradb # Rename the folder
cd defradb/cli/defradb
go install
```

## Getting Started
To get started with DefraDB, make sure you have the `defradb` cli installed locally, or access to a remote node.

Additionally, you most likely want to use a native GraphQL client, like GraphiQL, which can be downloaded as an Electron App [here](https://www.electronjs.org/apps/graphiql).

Setup a local DefraDB node with:
```
defradb start
```

This will start a node with the default settings (running at http://localhost:9181), and create a configuration file at $HOME/.defra/config.yaml. Where $HOME is your operating system user home directory.

Currently, DefraDB supports two storage engines; [BadgerDB](https://github.com/dgraph-io/badger), and an In-Memory store. By default, it uses BadgerDB, as it provides disk-backed persistent storage, unlike the In-Memory store. You can specify which engine to use with the `--store` option for the `start` command, or by editing the local config file.

If you are using BadgerDB, and you encounter the following error:
```
Failed to initiate database:Map log file. Path=.defradb/data/000000.vlog. Error=exec format error
```
It means terminal client doesn't support Mmap'ed files. This is common with older version of Ubuntu on Windows va WSL. Unfortuently, BadgerDB uses Mmap to interact with the filesystem, so you will need to use a terminal client which supports it.

Once your local environment is setup, you can test your connection with:
```
defradb client ping
```
which should respond with `Success!`

Once you've confirmed your node is running correctly, if you're using the GraphiQL client to interact with the database, then make sure you set the `GraphQL Endpoint` to `http://localhost:9181/graphql` and the `Method` to `GET`.

### Add a Schema type
To add a new schema type to the database, you can write the schema to a local file using the GraphQL SDL format, and submit that to the database like so:
```gql
# Add this to users.gql file
type user {
	name: String 
	age: Int 
	verified: Boolean 
	points: Float
}

# then run
defradb client schema add -f users.gql
```

This will register the type, build a dedicated collection, and generate the typed GraphQL endpoints for querying and mutation.

You can find more examples of schema type definitions in the [cli/defradb/examples](cli/defradb/examples) folder.

### Create a Document Instance
To create a new instance of a user type, submit the following query.
```gql
mutation {
    create_user(data: "{\"age\": 31, \"verified\": true, \"points\": 90, \"name\": \"Bob\"}") {
        _key
    }
}
```

This will respond with:
```json
{
  "data": [
    {
      "_key": "bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab",
    }
  ]
}
```
Here, the "_key" field is a unique identifier added to each and every document in a DefraDB node. It uses a combination of [UUIDs](https://en.wikipedia.org/wiki/Universally_unique_identifier) and [CIDs](https://docs.ipfs.io/concepts/content-addressing/)

### Query our documents
Once we have populated our local node with data, we can query that data. 
```gql
query {
  user {
    _key
    age
    name
    points
  }
}
```
This will query *all* the users, and return the fields `_key, age, name, points`. GraphQL queries only ever return the exact fields you request, there's no `*` selector like with SQL.

We can further filter our results by adding a `filter` argument to the query.
```gql
query {
  user(filter: {points: {_ge: 50}}) {
    _key
    age
    name
    points
  }
}
```
This will only return user documents which have a value for the points field *Greater Than or Equal to ( _ge )* `50`.

To see all the available query options, types, and functions please see the [Query Documenation](#query-documenation)

### Interact with Document Commits
Internally, DefraDB uses MerkleCRDTs to store data. MerkleCRDTs convert all mutations and updates a document has into a graph of changes; similar to Git. Moreover, the graph is a [MerkleDAG](https://docs.ipfs.io/concepts/merkle-dag/), which means all nodes are content identifiable with CIDs, and each node, references its parents CIDs.

To get the most recent commit in the MerkleDAG for a with a docKey of `bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab`, we can submit the following query:
```gql
query {
  latestCommits(dockey: "bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab") {
    cid
    delta
    height
    links {
      cid
      name
    }
  }
}
```
This will return a structure similar to the following, which contains the update payload that caused this new commit (delta), and any sub graph commits this references.
```json
{
  "data": [
    {
      "cid": "QmPqCtcCPNHoWkHLFvG4aKqDkLLzhVDAVEDSzEs38GHxoo",
      "delta": "pGNhZ2UYH2RuYW1lY0JvYmZwb2ludHMYWmh2ZXJpZmllZPU=",
      "height": 1,
      "links": [
        {
          "cid": "QmSom35RYVzYTE7nGsudvomv1pi9ffjEfSFsPZgQRM92v1",
          "name": "age"
        },
        {
          "cid": "QmYJrCcfMmfFp4JcbChLfMLCv8TSHjGwRVHUBgPazWxPga",
          "name": "name"
        },
        {
          "cid": "QmXLuVB5CCGqWcdQitingkfRxoVRLKh2jNcnX4UbYnW6Mk",
          "name": "points"
        },
        {
          "cid": "QmNRQwWjTBTDfAFUHkG8yuKmtbprYQtGs4jYxGJ5fCfXtn",
          "name": "verified"
        }
      ]
    }
  ]
}
```

Additionally, we can get *all* commits in a document MerkleDAG with `allCommits`, and lastly, we can get a specific commit, identified by a `cid` with the `commit` query, like so:
```gql
query {
  commit(cid: "QmPqCtcCPNHoWkHLFvG4aKqDkLLzhVDAVEDSzEs38GHxoo") {
    cid
    delta
    height
    links {
      cid
      name
    }
  }
}
```
Here, you can see we use the CID from the previous query to further explore the related nodes in the MerkleDAG.

This only scratches the surface of the DefraDB Query Language, see below for the entire language specification.

## Query Documenation
You can access the official DefraDB Query Language documentation online here: [https://hackmd.io/@source/BksQY6Qfw](https://hackmd.io/@source/BksQY6Qfw)

## CLI Documentation
You can find generated documentation for the shipped CLI interface [here](docs/cmd/defradb.md)

## Next Steps
The current early access release has much of the digial signatute, and identity work removed, until the cryptographic elements can be finalized.

The following will ship with the next release:
- schema type mutation/migration
- data syncronization between nodes
- grouping and aggregation on the query language
- additional CRDT type(s)
- and more. 

We will release a project board outlining the planned, and completed features of our roadmap.

## Licensing
Current DefraDB code is released under a combination of two licenses, the [Business Source License (BSL)](licenses/BSL.txt) and the [DefraDB Community License (DCL)](licenses/DCL.txt).

When contributing to a DefraDB feature, you can find the relevant license in the comments at the top of each file.

## Contributors
- John-Alan Simmons ([@jsimnz](https://github.com/jsimnz))

<br>

> The following documents are internal to the Source Team, if you wish to gain access, please reach out to us at [hello@source.network](mailto:hello@source.network)

## Futher Reading

### Technical Specification Doc
[https://hackmd.io/ZEwh3O15QN2u4p0cVoGbxw](https://hackmd.io/ZEwh3O15QN2u4p0cVoGbxw) (Private - Releasing soon)

### Design Doc
https://docs.google.com/document/d/10_7DiLFOOyTXBSM2wSsQxmcT9f1h44GBS3KIPDHs8n4/edit# (Private)
