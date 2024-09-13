---
sidebar_position: 1
title: Getting Started
slug: /
---

DefraDB is a user-centric database that prioritizes data ownership, personal privacy, and information security. Its data model, powered by the convergence of [MerkleCRDTs](https://arxiv.org/pdf/2004.00107.pdf) and the content-addressability of [IPLD](https://docs.ipld.io/), enables a multi-write-master architecture. It features [DQL](./references/query-specification/query-language-overview.md), a query language compatible with GraphQL but providing extra convenience. By leveraging peer-to-peer networking it can be deployed nimbly in novel topologies. Access control is determined by a relationship-based DSL, supporting document or field-level policies, secured by the SourceHub network. DefraDB is a core part of the [Source technologies](https://source.network/) that enable new paradigms of decentralized data and access-control management, user-centric apps, data trustworthiness, and much more.

DISCLAIMER: At this early stage, DefraDB does not offer access control or data encryption, and the default configuration exposes the database to the network. The software is provided "as is" and is not guaranteed to be stable, secure, or error-free. We encourage you to experiment with DefraDB and provide feedback, but please do not use it for production purposes until it has been thoroughly tested and developed.

## Install

Install `defradb` by [downloading an executable](https://github.com/sourcenetwork/defradb/releases) or building it locally using the [Go toolchain](https://golang.org/):

```shell
git clone git@github.com:sourcenetwork/defradb.git
cd defradb
make install
```

In the following sections, we assume that `defradb` is included in your `PATH`. If you installed it with the Go toolchain, use:

```shell
export PATH=$PATH:$(go env GOPATH)/bin
```

We recommend experimenting with queries using a native GraphQL client. Altair is a popular option - [download and install it](https://altairgraphql.dev/#download).

## Start

Start a node by executing `defradb start`. Keep the node running while going through the following examples.

Verify the local connection to the node works by executing `defradb client ping` in another terminal.

## Configuration

In this document, we use the default configuration, which has the following behavior:

- `~/.defradb/` is DefraDB's configuration and data directory
- `client` command interacts with the locally running node
- The GraphQL endpoint is provided at <http://localhost:9181/api/v0/graphql>

The GraphQL endpoint can be used with a GraphQL client (e.g., Altair) to conveniently perform requests (`query`, `mutation`) and obtain schema introspection.

## Add a schema type

Schemas are used to structure documents using a type system.

In the following examples, we'll be using a simple `User` schema type.

Add it to the database with the following command. By doing so, DefraDB generates the typed GraphQL endpoints for querying, mutation, and introspection.

```shell
defradb client schema add '
  type User {
    name: String 
    age: Int 
    verified: Boolean 
    points: Float
  }
'
```

Find more examples of schema type definitions in the [examples/schema/](https://github.com/sourcenetwork/defradb/examples/schema/) folder.

## Create a document

Submit a `mutation` request to create a document of the `User` type:

```shell
defradb client query '
  mutation {
      create_User(input: {age: 31, verified: true, points: 90, name: "Bob"}) {
          _docID
      }
  }
'
```

Expected response:

```json
{
  "data": [
    {
      "_docID": "bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab",
    }
  ]
}
```

`_docID` is the document's key, a unique identifier of the document, determined by its schema and initial data.

## Query documents

Once you have populated your node with data, you can query it:

```shell
defradb client query '
  query {
    User {
      _docID
      age
      name
      points
    }
  }
'
```

This query obtains *all* users and returns their fields `_docID, age, name, points`. GraphQL queries only return the exact fields requested.

You can further filter results with the `filter` argument.

```shell
defradb client query '
  query {
    User(filter: {points: {_ge: 50}}) {
      _docID
      age
      name
      points
    }
  }
'
```

This returns only user documents which have a value for the `points` field *Greater Than or Equal to* (`_ge`) 50.

## Obtain document commits

DefraDB's data model is based on [MerkleCRDTs](./guides/merkle-crdt.md). Each document has a graph of all of its updates, similar to Git. The updates are called `commits` and are identified by `cid`, a content identifier. Each references its parents by their `cid`s.

To get the most recent commits in the MerkleDAG for the document identified as `bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab`:

```shell
defradb client query '
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
'
```

It returns a structure similar to the following, which contains the update payload that caused this new commit (`delta`) and any subgraph commits it references.

```json
{
  "data": {
    "latestCommits": [
      {
        "cid": "bafybeifhtfs6vgu7cwbhkojneh7gghwwinh5xzmf7nqkqqdebw5rqino7u",
        "delta": "pGNhZ2UYH2RuYW1lY0JvYmZwb2ludHMYWmh2ZXJpZmllZPU=",
        "height": 1,
        "links": [
          {
            "cid": "bafybeiet6foxcipesjurdqi4zpsgsiok5znqgw4oa5poef6qtiby5hlpzy",
            "name": "age"
          },
          {
            "cid": "bafybeielahxy3r3ulykwoi5qalvkluojta4jlg6eyxvt7lbon3yd6ignby",
            "name": "name"
          },
          {
            "cid": "bafybeia3tkpz52s3nx4uqadbm7t5tir6gagkvjkgipmxs2xcyzlkf4y4dm",
            "name": "points"
          },
          {
            "cid": "bafybeia4off4javopmxcdyvr6fgb5clo7m5bblxic5sqr2vd52s6khyksm",
            "name": "verified"
          }
        ]
      }
    ]
  }
}
```

Obtain a specific commit by its content identifier (`cid`):

```shell
defradb client query '
  query {
    commits(cid: "bafybeifhtfs6vgu7cwbhkojneh7gghwwinh5xzmf7nqkqqdebw5rqino7u") {
      cid
      delta
      height
      links {
        cid
        name
      }
    }
  }
'
```

## DefraDB Query Language (DQL)

DQL is compatible with GraphQL but features various extensions.

Read the [Query specification](./references/query-specification/query-language-overview.md) to discover filtering, ordering, limiting, relationships, variables, aggregate functions, and other useful features.


## Peer-to-peer data synchronization

DefraDB leverages peer-to-peer networking for data exchange, synchronization, and replication of documents and commits.

When starting a node for the first time, a key pair is generated and stored in its "root directory" (`~/.defradb/` by default).

Each node has a unique `Peer ID` generated from its public key. This ID allows other nodes to connect to it.

There are two types of peer-to-peer relationships supported: **pubsub** peering and **replicator** peering.

Pubsub peering *passively* synchronizes data between nodes by broadcasting *Document Commit* updates to the topic of the commit's document key. Nodes need to be listening on the pubsub channel to receive updates. This is for when two nodes *already* have share a document and want to keep them in sync.

Replicator peering *actively* pushes changes from a specific collection *to* a target peer.

### Pubsub example

Pubsub peers can be specified on the command line using the `--peers` flag, which accepts a comma-separated list of peer [multiaddresses](https://docs.libp2p.io/concepts/addressing/). For example, a node at IP `192.168.1.12` listening on 9000 with Peer ID `12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B` would be referred to using the multiaddress `/ip4/192.168.1.12/tcp/9000/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B`.

Let's go through an example of two nodes (*nodeA* and *nodeB*) connecting with each other over pubsub, on the same machine.

Start *nodeA* with a default configuration:

```shell
defradb start
```

Obtain the Peer ID from its console output. In this example, we use `12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B`, but locally it will be different.

For *nodeB*, we provide the following configuration:

```shell
defradb start --rootdir ~/.defradb-nodeB --url localhost:9182 --p2paddr /ip4/0.0.0.0/tcp/9172 --tcpaddr /ip4/0.0.0.0/tcp/9162 --peers /ip4/0.0.0.0/tcp/9171/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B
```

About the flags:

- `--rootdir` specifies the root dir (config and data) to use
- `--url` is the address to listen on for the client HTTP and GraphQL API
- `--p2paddr` is the multiaddress for the p2p networking to listen on
- `--tcpaddr` is the multiaddress for the gRPC server to listen on
- `--peers` is a comma-separated list of peer multiaddresses

This starts two nodes and connects them via pubsub networking.

### Collection subscription example

It is possible to subscribe to updates on a given collection by using its ID as the pubsub topic. The ID of a collection is found as the field `schemaVersionID` in one of its documents. Here we use the collection ID of the `User` type we created above. After setting up 2 nodes as shown in the [Pubsub example](#pubsub-example) section, we can subscribe to collections updates on *nodeA* from *nodeB* by using the `rpc p2pcollection` command:

```shell
defradb client rpc p2pcollection add --url localhost:9182 bafkreibpnvkvjqvg4skzlijka5xe63zeu74ivcjwd76q7yi65jdhwqhske
```

Multiple collection IDs can be added at once.

```shell
defradb client rpc p2pcollection add --url localhost:9182 <collection1ID> <collection2ID> <collection3ID>
```

### Replicator example

Replicator peering is targeted: it allows a node to actively send updates to another node. Let's go through an example of *nodeA* actively replicating to *nodeB*:

Start *nodeA*:

```shell
defradb start
```

In another terminal, add this example schema to it:

```shell
defradb client schema add '
  type Article {
    content: String
    published: Boolean
  }
'
```

Start (or continue running from above) *nodeB*, that will be receiving updates:

```shell
defradb start --rootdir ~/.defradb-nodeB --url localhost:9182 --p2paddr /ip4/0.0.0.0/tcp/9172 --tcpaddr /ip4/0.0.0.0/tcp/9162
```

Here we *do not* specify `--peers` as we will manually define a replicator after startup via the `rpc` client command.

In another terminal, add the same schema to *nodeB*:

```shell
defradb client schema add --url localhost:9182 '
  type Article {
    content: String
    published: Boolean
  }
'
```

Set *nodeA* to actively replicate the "Article" collection to *nodeB*:

```shell
defradb client rpc addreplicator "Article" /ip4/0.0.0.0/tcp/9172/p2p/
defradb client rpc replicator set -c "Article" /ip4/0.0.0.0/tcp/9172/p2p/<peerID_of_nodeB>

```

As we add or update documents in the "Article" collection on *nodeA*, they will be actively pushed to *nodeB*. Note that changes to *nodeB* will still be passively published back to *nodeA*, via pubsub.


## Securing the HTTP API with TLS

By default, DefraDB will expose its HTTP API at `http://localhost:9181/api/v0`. It's also possible to configure the API to use TLS with self-signed certificates or Let's Encrypt.

To start defradb with self-signed certificates placed under `~/.defradb/certs/` with `server.key`
being the public key and `server.crt` being the private key, just do:
```shell
defradb start --tls
```

The keys can be generated with your generator of choice or with `make tls-certs`.

Since the keys should be stored within the DefraDB data and configuration directory, the recommended key generation command is `make tls-certs path="~/.defradb/certs"`.

If not saved under `~/.defradb/certs` then the public (`pubkeypath`) and private (`privkeypaths`) key paths need to be explicitly defined in addition to the `--tls` flag or `tls` set to `true` in the config.

Then to start the server with TLS, using your generated keys in custom path:
```shell
defradb start --tls --pubkeypath ~/path-to-pubkey.key --privkeypath ~/path-to-privkey.crt

```

DefraDB also comes with automatic HTTPS for deployments on the public web. To enable HTTPS,
 deploy DefraDB to a server with both port 80 and port 443 open. With your domain's DNS A record
 pointed to the IP of your server, you can run the database using the following command:
```shell
sudo defradb start --tls --url=your-domain.net --email=email@example.com
```
Note: `sudo` is needed above for the redirection server (to bind port 80).

A valid email address is necessary for the creation of the certificate, and is important to get notifications from the Certificate Authority - in case the certificate is about to expire, etc.


## Conclusion

This gets you started to use DefraDB! Read on the documentation website for guides and further information.
