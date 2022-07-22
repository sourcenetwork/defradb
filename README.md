[![codecov](https://codecov.io/gh/sourcenetwork/defradb/branch/develop/graph/badge.svg?token=RHAORX13PA)](https://codecov.io/gh/sourcenetwork/defradb)

<p align="center">
<img height="120px" src="docs/DefraDB_Full-v2-cropped.png">
</p>

DefraDB is a peer-to-peer edge document database redefining promises of data ownership, personal privacy, and information security, around the user. It features a GraphQL-compatible query language called DQL. Its data model, enabled by [MerkleCRDTs](https://arxiv.org/pdf/2004.00107.pdf), makes possible a multi-write-master architecture. It is the core data system for the [Source](https://source.network/) ecosystem. It is built with technologies like [IPLD](https://docs.ipld.io/) and [libP2P](https://libp2p.io/), and featuring Web3 and semantic properties.

Read the [Technical Overview](https://docsend.com/view/zwgut89ccaei7e2w/d/bx4vu9tj62bewenu) and documentation on [docs.source.network](https://docs.source.network/).


## Early Access

DefraDB is currently in a *Early Access Alpha* program, and is not yet ready for production deployments. Please email the [Source](https://source.network/) team at [hello@source.network](mailto:hello@source.network) for support with your use-case and deployment.


## Install

Install `defradb` by [downloading an executable binary](https://github.com/sourcenetwork/defradb/releases), or building it locally using the [Go toolchain](https://golang.org/):
```sh
git clone git@github.com:sourcenetwork/defradb.git
make install
```

It is recommended to play around with queries using a native GraphQL client. GraphiQL is a popular option - [download it](https://www.electronjs.org/apps/graphiql).


## Start

`defradb start` spins up a node. Keep it running to perform the following examples.

Verify you are properly connected to the node using `defradb client ping`.

By default, `~/.defradb/` is the configuration and data directory, and a GraphQL endpoint is provided at http://localhost:9181/api/v0/graphql.

Connect a GraphQL client (e.g. GraphiQL) to the endpoint to conveniently obtain introspection and perform requests (`query`, `mutation`).


## Add a schema type

Schemas are used to structure documents using a type system.

In the following examples we'll be using the following `user` schema type. Write it to the `users.gql` local file in GraphQL SDL format:
```gql
type user {
	name: String 
	age: Int 
	verified: Boolean 
	points: Float
}
```

Then add it to the database:
```
defradb client schema add -f users.gql
```

Adding a schema will generate the typed GraphQL endpoints for querying and mutation.

Find more examples of schema type definitions in the [examples/schema/](examples/schema/) folder.


## Create a document instance

To create an instance of a user type, submit the following request:
```gql
mutation {
    create_user(data: "{\"age\": 31, \"verified\": true, \"points\": 90, \"name\": \"Bob\"}") {
        _key
    }
}
```

Submit a request via a GraphQL client, or using:

```
defradb client query 'insert query here'
```

It will respond:
```json
{
  "data": [
    {
      "_key": "bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab",
    }
  ]
}
```

The "_key" field is a unique identifier added to each and every document in a DefraDB node. It uses a combination of [UUIDs](https://en.wikipedia.org/wiki/Universally_unique_identifier) and [CIDs](https://docs.ipfs.io/concepts/content-addressing/).


## Query documents

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

This query obtains *all* users, and return the fields `_key, age, name, points`. GraphQL queries only ever return the exact fields you request, there's no `*` selector like with SQL.

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

This will return only user documents which have a value for the `points` field *Greater Than or Equal to* (`_ge`) 50.


## Obtain document commits

Internally, data is handled by MerkleCRDTs, which convert all mutations and updates a document has into a graph of changes; similar to Git. The graph is a [MerkleDAG](https://docs.ipfs.io/concepts/merkle-dag/), which means all nodes are content-identifiable with, and each node references its parents CIDs.

To get the most recent commit in the MerkleDAG for a document with a DocKey of `bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab`, submit:
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

It returns a structure similar to the following, which contains the update payload that caused this new commit (delta), and any subgraph commits it references.
```json
{
  "data": [
    {
      "cid": "bafybeidembipteezluioakc2zyke4h5fnj4rr3uaougfyxd35u3qzefzhm",
      "delta": "pGNhZ2UYH2RuYW1lY0JvYmZwb2ludHMYWmh2ZXJpZmllZPU=",
      "height": 1,
      "links": [
        {
          "cid": "bafybeieelb43ol5e5jiick2p7k4p577ph72ecwcuowlhbops4hpz24zhz4",
          "name": "age"
        },
        {
          "cid": "bafybeigwjkwz7eobh6pqjal4yfpsahnv74cxedxqnmhnmp3iojoc4xs25i",
          "name": "name"
        },
        {
          "cid": "bafybeigr72knx43euvxz34sqipjrfpcgibr7uuihko4aqxqutkkmzhqm24",
          "name": "points"
        },
        {
          "cid": "bafybeig653zyvev625vkn5kveuhyzqnychvoqhyx52pznbre227olkzkpi",
          "name": "verified"
        }
      ]
    }
  ]
}
```

Obtain a specific commit by its CID:
```gql
query {
  commit(cid: "bafybeidembipteezluioakc2zyke4h5fnj4rr3uaougfyxd35u3qzefzhm") {
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


## Query language documentation

Read the full DefraDB Query Language documentation on [docs.source.network](https://docs.source.network/query-specification/query-language-overview).

You will discover about filtering, ordering, limiting, relationships, variables, aggregate functions, and further useful features.


## Peer-to-peer data synchronization
DefraDB uses P2P networking for nodes to exchange, synchronize, and replicate documents and commits.

When starting a node for the first time, a key pair is generated and stored it in its "root folder" (commonly `~/.defradb/`).

Each node has a unique `Peer ID` generated based on the public key, which is printed to the console during startup. This ID allows other nodes to connect to it.

There are two types of peer-to-be relationships nodes support: **pubsub** peering and **replicator** peering.

Pubsub peering *passively* synchronizes data between nodes by broadcasting Document Commit updates with the document `DocKey` as the topic. Nodes need to already be listening on the pubsub channel to receive updates. This is for when two nodes *already* have a shared document and want to keep both their changes in sync with one another.

Replicator peering *actively* pushes changes from a specific collection *to* a target peer.


### Pubsub example

Pubsub peers can be specified on the command line using the `--peers` flag which accepts a comma-separated list of peer [multiaddress](https://docs.libp2p.io/concepts/addressing/). For example, a node at `192.168.1.12` listening on 9000 with Peer ID `12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B` would be referred to using the multiaddress  `/ip4/192.168.1.12/tcp/9000/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B`.

Let's go through an example of two nodes (*nodeA* and *nodeB*) connecting with each other over pubsub, on the same machine.

Start *nodeA* with a default configuration:
```
defradb start
```

Obtain the Peer ID from its console output. In this example, we assume `12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B`, but locally it will be different.

For *nodeB*, we provide the following configuration:
```
defradb start --rootdir ~/.defradb-nodeB --url localhost:9182 --p2paddr /ip4/0.0.0.0/tcp/9172 --tcpaddr /ip4/0.0.0.0/tcp/9162 --peers /ip4/0.0.0.0/tcp/9171/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B
```

About the flags:

- `--rootdir` specifies the root dir (config and data) to use
- `--url` is the address to listen on for the client HTTP and GraphQL API
- `--p2paddr` is the multiaddress for the p2p networking to listen on
- `--tcpaddr` is the multiaddress for the gRPC server to listen on
- `--peers` is a comma-separated list of peer multiaddresses

This starts two nodes and connect them via pubsub networking.


### Replicator example

Replicator peering is established in one direction. For example, a *nodeA* can be given a *nodeB* to actively send updates to, but *nodeB* won't send updates in return. However, nodes broadcast updates of documents over document-specific pubsub topics, therefore *nodeB* while it won't replicate directly to *nodeA*, it will *passively*.

Let's go through an example of *nodeA* actively replicating to *nodeB*:

Start *nodeA* **and** define a collection. We will use it as leader.
```
defradb start
TBD
```

Start *nodeB* as follower.
```
defradb start --rootdir ~/.defradb-nodeB --url localhost:9182 --p2paddr /ip4/0.0.0.0/tcp/9172 --tcpaddr /ip4/0.0.0.0/tcp/9162
```

Notice how we *do not* specify `--peers` as we will manually define a replicator after startup via the `rpc` client command.

On *nodeB*, in another terminal, add an example schema:
```shell
defradb client schema add --url localhost:9182 '
  type user {
    name: String 
    age: Int 
    verified: Boolean 
    points: Float
  }
'
```

On *nodeA*, add the same schema and set *nodeB* as target replicator peer:
```
defradb client schema add '
  type user {
    name: String 
    age: Int 
    verified: Boolean 
    points: Float
  }
'
defradb client rpc add-replicator --addr 0.0.0.0:9162 user <nodeB_peer_address>
```

With this, as we add documents to *nodeA*, they will be actively pushed to *nodeB*, and when we make changes to *nodeB* they will be passively published back to *nodeA*.


## Licensing

DefraDB's code is released under the [Business Source License (BSL)](licenses/BSL.txt). It grants you the right to copy, modify, create derivative works, redistribute, and make non-production use of it. For additional uses, such as deploying in production on a private network, please contact license@source.network for a licensing agreement. Each dated version of the license turns into an Apache License v2.0 after 4 years. Please read the complete license before usage.


## Contributors

- John-Alan Simmons ([@jsimnz](https://github.com/jsimnz))
- Andrew Sisley ([@AndrewSisley](https://github.com/AndrewSisley))
- Shahzad Lone ([@shahzadlone](https://github.com/shahzadlone))
- Orpheus Lummis ([@orpheuslummis](https://github.com/orpheuslummis))
- Fred Carle ([@fredcarle](https://github.com/fredcarle))

You are invited to contribute to DefraDB. Follow the [Contributing guide](./CONTRIBUTING.md) to get started.
