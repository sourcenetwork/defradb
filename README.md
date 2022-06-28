[![codecov](https://codecov.io/gh/sourcenetwork/defradb/branch/develop/graph/badge.svg?token=RHAORX13PA)](https://codecov.io/gh/sourcenetwork/defradb)

<p align="center">
<img height="120px" src="docs/DefraDB_Full-v2-cropped.png">
</p>

#
The DefraDB is a Peer-to-Peer Edge Database, with the interface of a NoSQL Document Store. DefraDB's data model is backed by [MerkleCRDTs](https://arxiv.org/pdf/2004.00107.pdf) for a multi write-master architecture. It is the core data storage system for the [Source](https://source.network/) Ecosystem, built with [IPFS](https://ipfs.io/) technologies like [IPLD](https://docs.ipld.io/) and [LibP2P](https://libp2p.io/) and featuring Semantic web3 properties. You can read the [Technical Overview here](https://docsend.com/view/mczj7ic4i3kqpq7s).

Table of Contents
=================

* [Early Access](#early-access)
* [Installation](#installation)
	* [Build](#build)
* [Getting Started](#getting-started)
	* [Add a Schema type](#add-a-schema-type)
	* [Create a Document Instance](#create-a-document-instance)
	* [Query our documents](#query-our-documents)
	* [Interact with Document Commits](#interact-with-document-commits)
* [Query Documentation](#query-documenation)
* [CLI Documentation](#cli-documentation)
* [Next Steps](#next-steps)
* [Licensing](#licensing)
* [Contributors](#contributors)
* [Further Reading](#futher-reading)
	* [Technical Specification Doc](#technical-specification-doc)
	* [Design Doc](#design-doc)

## Early Access

DefraDB is currently in a *Early Access Alpha* program, and is not yet ready for production deployments. Please reach out to the team at [Source](https://source.network/) by emailing [hello@source.network](mailto:hello@source.network) for support with your use-case and deployment.

## Installation

To install a DefraDB node, you can download the pre-compiled binaries available on the releases page, or you can compile it yourself if you have a local [Go Toolchain](https://golang.org/) installed.

### Build

Build dependencies:
```
make deps

```
Run using Make:
```
make
```

Install using Make:
```
make install
```

Build the binary inside the `build/` folder:
```
make build
```

Build and Start DefraDB:
```
make start
```

Install Manually:
```
go install github.com/sourcenetwork/defradb/cli/defradb
```

Build and Install Manually:
```
git clone git@github.com:sourcenetwork/defradb.git
cd defradb/cli/defradb
go install
```

### Test

Run the tests:
```
make test
```

Run the benchmark tests:
```
make test:bench
```

### Lint

Run the linter (make sure to have had run `make deps` before or have `golangci-lint` installed):
```
make lint
```

Find all occurrences where the linter errors are explicitly being suppressed (need to resolve these):
```
make lint:todo
```

List all enabled and disabled linters:
```
make lint:list
```

## Getting Started

To get started with DefraDB, make sure you have the `defradb` cli installed locally, or access to a remote node.

Additionally, you most likely want to use a native GraphQL client, like GraphiQL, which can be downloaded as an Electron App [here](https://www.electronjs.org/apps/graphiql).

Setup a local DefraDB node with:
```
make start
```

OR by manually building and then running:

```
defradb start
```

This will start a node with the default settings (running at http://localhost:9181), and create a configuration file at $HOME/.defra/config.yaml. Where $HOME is your operating system user home directory.

Currently, DefraDB supports two storage engines; [BadgerDB](https://github.com/dgraph-io/badger), and an In-Memory store. By default, it uses BadgerDB, as it provides disk-backed persistent storage, unlike the In-Memory store. You can specify which engine to use with the `--store` option for the `start` command, or by editing the local config file.

If you are using BadgerDB, and you encounter the following error:
```
Failed to initiate database:Map log file. Path=.defradb/data/000000.vlog. Error=exec format error
```
It means terminal client doesn't support Mmap'ed files. This is common with older version of Ubuntu on Windows va WSL. Unfortunately, BadgerDB uses Mmap to interact with the filesystem, so you will need to use a terminal client which supports it.

Once your local environment is setup, you can test your connection with:
```
defradb client ping
```
which should respond with `Success!`

Once you've confirmed your node is running correctly, if you're using the GraphiQL client to interact with the database, then make sure you set the `GraphQL Endpoint` to `http://localhost:9181/api/v0/graphql`.

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

You can find more examples of schema type definitions in the [examples/schema](examples/schema) folder.

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
      "cid": "bafkreibbbtps3cki27zs5w24djizgjfuydqsdgtrz773sfpqechiih5ose",
      "delta": "pGNhZ2UYH2RuYW1lY0JvYmZwb2ludHMYWmh2ZXJpZmllZPU=",
      "height": 1,
      "links": [
        {
          "cid": "bafybeiccmckwnoe3ib3ebybcb6fdyawrji4z4wg2trwswjbfmkeyhdkd4y",
          "name": "age"
        },
        {
          "cid": "bafybeieud2zx7evz47vtujk4pyhexcaankozyfhkd4tyjoyp73eeq2xak4",
          "name": "name"
        },
        {
          "cid": "bafybeiefzhsaukwg6dazgyztulstewns2f6a6c6hku4egso2cyowxrxcd4",
          "name": "points"
        },
        {
          "cid": "bafybeiabhbk6omcebl52bwxjxbcabpe2ajc22yxrav7mq5f26v7ds76j2m",
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

## Query Documentation

You can access the official DefraDB Query Language documentation online here: [https://hackmd.io/@source/BksQY6Qfw](https://hackmd.io/@source/BksQY6Qfw)

## Peer-to-Peer Data Synchronization
DefraDB has a native P2P network builtin to each node, allowing them to exchange, synchronize, and replicate documents and commits.

The P2P network uses a combination of server to server gRPC commands, gossip based pub-sub network, and a shared Distributed Hash Table, all powered by
[LibP2P](https://libp2p.io/).

Unless specifying `--no-p2p` option when running `start` the default behaviour for a DefraDB node is to initialize the P2P network stack.

When you start a node for the first time, DefraDB will auto generate a private key pair and store it in the `data` folder specified in the config or `--data` CLI option. Each node has a unique `Peer ID` generated based on the public key, which is printed to the console during startup.

You'll see a printed line: `Created LibP2P host with Peer ID XXX` where `XXX` is your node's `Peer ID`. This is important to know if we want other nodes to connect to this node.

There are two types of relationships a given DefraDB node can establish with another peer, which is a pubsub peer or a replicator peer. 

Pubsub peers can be specified on the command line with `--peers` which accepts a comma-separated list of peer [MultiAddress](https://docs.libp2p.io/concepts/addressing/). Which takes the form of `/ip4/IP_ADDRESS/tcp/PORT/p2p/PEER_ID`.

> If a node is listening on port 9000 with the IP address `192.168.1.12` and a Peer ID of `12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B` then the fully quantified multi address is `/ip4/192.168.1.12/tcp/9000/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B`.

Pubsub nodes *passively* synchronize data between nodes by broadcasting Document Commit updates over the pubsub channel with the document `DocKey` as the topic. This requires nodes to already be listening on this pubsub channel to receive updates for. This is used when two nodes *already* have a shared document, and want to keep both their changes in sync with one another.

Replicator nodes are specified using the CLI `rpc` command after a node has already started with `defradb rpc add-replicator <collection> <peer_multiaddress>`.

Replicator nodes *actively* push changes from the specific collection *to* the target peer.

> Note: Replicator nodes are initially established in one direction by default, so Node1 *actively replicates* to Node2 but not vice-versa. However, nodes will always broadcast their updates over the document specific pubsub topic so while Node2 doesn't replicate directly to Node1, it will *passively*.

### PubSub Example

Let's construct a simple example of two nodes (node1 & node2) connecting to one another over the pubsub network on the same machine.

On Node1 start a regular node with all the defaults:
```
defradb start
```

Make sure to get the `Peer ID` from the console output. Let's assume its `12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B`.

One Node2 we need to change some of the default config options if we are running on the same machine.
```
defradb start --data $HOME/.defradb/data-node2 --p2paddr /ip4/0.0.0.0/tcp/9172 --url localhost:9182 --peers /ip4/0.0.0.0/tcp/9171/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B
```

Let's break this down
- `--data` specifies the data folder
- `--p2paddr` is the multiaddress to listen on for the p2p network (default is port 9171)
- `--url` is the HTTP address to listen on for the client HTTP and GraphQL API.
- `--peers`  is a comma-separated list of peer multiaddresses. This will be our first node we started, with the default config options.

This will startup two nodes, connect to each other, and establish the P2P gossib pubsub network. 

### Replicator Example

Let's construct a simple example of Node1 *replicating* to Node2.

Node1 is the leader, let's startup the node **and** define a collection.
```
defradb start
```

On Node2 let's startup a node
```
defradb start --data $HOME/.defradb/data-node2 --p2paddr /ip4/0.0.0.0/tcp/9172 --url localhost:9182
```

You'll notice we *don't* specify the `--peers` option as we will be manually defining a replicator after startup via the `rpc` client command.

On Node2 in another terminal run
```
defradb client schema add -f <your_schema.gql> # Add a collection with a schema
```

On Node1, run:
```
defradb client schema add -f <your_schema.gql> # Add the same collection as Node2
defradb rpc add-replicator <collection_name> <node2_peer_address> # Set Node2 as our target replicator peer
```

Now if we add documents to Node1, they will be *actively* pushed to Node2, and if we make changes to Node2 they will be *passively* published back to Node1. See the below section "CLI Documentation" for help with these commands


## CLI Documentation
You can always access the CLI help docs with the `defradb help` or `defradb <subcommand> help`.

You can also find generated markdown documentation for the shipped CLI interface [here](docs/cmd/defradb.md).

## Next Steps

The current early access release has much of the digital signature, and identity work removed, until the cryptographic elements can be finalized.

The following will ship with the next release:
- schema type mutation/migration
- data synchronization between nodes
- grouping and aggregation on the query language
- additional CRDT type(s)
- and more. 

We will release a project board outlining the planned, and completed features of our roadmap.

## Licensing

Current DefraDB code is released under a combination of two licenses, the [Business Source License (BSL)](licenses/BSL.txt) and the [DefraDB Community License (DCL)](licenses/DCL.txt).

When contributing to a DefraDB feature, you can find the relevant license in the comments at the top of each file.

## Contributors
- John-Alan Simmons ([@jsimnz](https://github.com/jsimnz))
- Andrew Sisley ([@AndrewSisley](https://github.com/AndrewSisley))
- Shahzad Lone ([@shahzadlone](https://github.com/shahzadlone))
- Orpheus Lummis ([@orpheuslummis](https://github.com/orpheuslummis))
- Fred Carle ([@fredcarle](https://github.com/fredcarle))

<br>

> The following documents are internal to the Source Team, if you wish to gain access, please reach out to us at [hello@source.network](mailto:hello@source.network)

## Further Reading

### Technical Specification Doc

[https://hackmd.io/ZEwh3O15QN2u4p0cVoGbxw](https://hackmd.io/ZEwh3O15QN2u4p0cVoGbxw) (Private - Releasing soon)

### Design Doc

https://docs.google.com/document/d/10_7DiLFOOyTXBSM2wSsQxmcT9f1h44GBS3KIPDHs8n4/edit# (Private)
