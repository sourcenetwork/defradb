<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/DefraDB_White.svg">
    <img height="120px" width="374px" alt="DefraDB" src="docs/DefraDB_Full.svg">
  </picture>
</p>

DefraDB is a user-centric database that prioritizes data ownership, personal privacy, and information security. Its data model, powered by the convergence of [MerkleCRDTs](https://arxiv.org/pdf/2004.00107.pdf) and the content-addressability of [IPLD](https://docs.ipld.io/), enables a multi-write-master architecture. It features [DQL](https://docs.source.network/query-specification/query-language-overview), a query language compatible with GraphQL but providing extra convenience. By leveraging peer-to-peer networking and [WebAssembly](https://en.wikipedia.org/wiki/WebAssembly), it can be deployed nimbly in novel topologies. Access control is determined by a relationship-based DSL, supporting document or field-level policies, secured by the SourceHub network. DefraDB is a core part of the [Source technologies](https://source.network/) that enable new paradigms of decentralized data and access-control management, user-centric apps, data trustworthiness, and much more.

Read the [Technical Overview](https://docsend.com/view/zwgut89ccaei7e2w/d/bx4vu9tj62bewenu) and documentation on [docs.source.network](https://docs.source.network/).

## Table of Contents

- [Early Access](#early-access)
- [Install](#install)
- [Start](#start)
- [Configuration](#configuration)
- [Add a schema type](#add-a-schema-type)
- [Create a document instance](#create-a-document-instance)
- [Query documents](#query-documents)
- [Obtain document commits](#obtain-document-commits)
- [DefraDB Query Language (DQL)](#defradb-query-language-dql)
- [Peer-to-peer data synchronization](#peer-to-peer-data-synchronization)
  - [Pubsub example](#pubsub-example)
  - [Collection subscription example](#collection-subscription-example)
  - [Replicator example](#replicator-example)
- [Securing the HTTP API with TLS](#securing-the-http-api-with-tls)
- [Licensing](#licensing)
- [Contributors](#contributors)



DISCLAIMER: At this early stage, DefraDB does not offer access control or data encryption, and the default configuration exposes the database to the network. The software is provided "as is" and is not guaranteed to be stable, secure, or error-free. We encourage you to experiment with DefraDB and provide feedback, but please do not use it for production purposes until it has been thoroughly tested and developed.

## Install

Install `defradb` by [downloading an executable binary](https://github.com/sourcenetwork/defradb/releases) or building it locally using the [Go toolchain](https://golang.org/):

```sh
git clone git@github.com:sourcenetwork/defradb.git
cd defradb
make install
```

In the following sections, we assume that `defradb` is included in your `PATH`. If you installed it with the Go toolchain, use:

```sh
export PATH=$PATH:$(go env GOPATH)/bin
```

We recommend experimenting with queries using a native GraphQL client. GraphiQL is a popular option - [download and install it](https://www.electronjs.org/apps/graphiql).

## Start

Start a node by executing `defradb start`. Keep the node running while going through the following examples.

Verify the local connection to the node works by executing `defradb client ping` in another terminal.

## Configuration

In this document, we use the default configuration, which has the following behavior:

- `~/.defradb/` is DefraDB's configuration and data directory
- `client` command interacts with the locally running node
- The GraphQL endpoint is provided at http://localhost:9181/api/v0/graphql

The GraphQL endpoint can be used with a GraphQL client (e.g., GraphiQL) to conveniently perform requests (`query`, `mutation`) and obtain schema introspection.

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

Find more examples of schema type definitions in the [examples/schema/](examples/schema/) folder.

## Create a document instance

Submit a `mutation` request to create an instance of the `User` type:

```shell
defradb client query '
  mutation {
      create_User(data: "{\"age\": 31, \"verified\": true, \"points\": 90, \"name\": \"Bob\"}") {
          _key
      }
  }
'
```

Expected response:

```json
{
  "data": [
    {
      "_key": "bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab",
    }
  ]
}
```

The document's `_key` is a unique identifier added to each document in a DefraDB node, determined by its schema and initial data.

## Query documents

Once you have populated your local node with data, you can query it:

```shell
defradb client query '
  query {
    User {
      _key
      age
      name
      points
    }
  }
'
```

This query obtains *all* users and returns their fields `_key, age, name, points`. GraphQL queries only return the exact fields requested.

You can further filter results with the `filter` argument.

```shell
defradb client query '
  query {
    User(filter: {points: {_ge: 50}}) {
      _key
      age
      name
      points
    }
  }
'
```

This returns only user documents which have a value for the `points` field *Greater Than or Equal to* (`_ge`) 50.

## Obtain document commits

DefraDB's data model is based on [MerkleCRDTs](https://arxiv.org/pdf/2004.00107.pdf). Each document has a graph of all of its updates, similar to Git. The updates are called `commit`s and are identified by `cid`, a content identifier. Each references its parents by their `cid`s.

To get the most recent commit in the MerkleDAG for the document identified as `bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab`:

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

Obtain a specific commit by its content identifier (`cid`):

```gql
defradb client query '
  query {
    commits(cid: "bafybeidembipteezluioakc2zyke4h5fnj4rr3uaougfyxd35u3qzefzhm") {
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

Read its documentation at [docs.source.network](https://docs.source.network/query-specification/query-language-overview) to discover its filtering, ordering, limiting, relationships, variables, aggregate functions, and other useful features.

## Peer-to-peer data synchronization
DefraDB leverages peer-to-peer networking for data exchange, synchronization, and replication of documents and commits.

When starting a node for the first time, a key pair is generated and stored in its "root directory" (commonly `~/.defradb/`).

Each node has a unique `Peer ID` generated from its public key. This ID allows other nodes to connect to it.

There are two types of peer-to-peer relationships nodes support: **pubsub** peering and **replicator** peering.

Pubsub peering *passively* synchronizes data between nodes by broadcasting Document Commit updates with the document key (`DocKey`) as the topic. Nodes need to already be listening on the pubsub channel to receive updates. This is for when two nodes *already* have a shared document and want to keep both their changes in sync with one another.

Replicator peering *actively* pushes changes from a specific collection *to* a target peer.

### Pubsub example

Pubsub peers can be specified on the command line using the `--peers` flag, which accepts a comma-separated list of peer [multiaddresses](https://docs.libp2p.io/concepts/addressing/). For example, a node at `192.168.1.12` listening on 9000 with Peer ID `12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B` would be referred to using the multiaddress `/ip4/192.168.1.12/tcp/9000/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B`.

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

This starts two nodes and connects them via pubsub networking.

### Collection subscription example

It is possible to subscribe to updates on a given collection by using its ID as the pubsub topic. After setting up 2 nodes as shown in the [Pubsub example](#pubsub-example) section, we can subscribe to collections updates on *nodeA* from *nodeB* by using the `rpc p2pcollection` command:

```shell
defradb client rpc p2pcollection add --url localhost:9182 <collectionID>
```

Multiple collection IDs can be added at once.

```shell
defradb client rpc p2pcollection add --url localhost:9182 <collection1ID> <collection2ID> <collection3ID>
```

### Replicator example

Replicator peering is established in one direction. For example, a *nodeA* can be given a *nodeB* to actively send updates to, but *nodeB* won't send updates in return. However, nodes broadcast updates of documents over document-specific pubsub topics, therefore *nodeB* while it won't replicate directly to *nodeA*, it will *passively*.

Let's go through an example of *nodeA* actively replicating to *nodeB*:

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

Start *nodeB*, that will be receiving updates:

```shell
defradb start --rootdir ~/.defradb-nodeB --url localhost:9182 --p2paddr /ip4/0.0.0.0/tcp/9172 --tcpaddr /ip4/0.0.0.0/tcp/9162
```

Notice how we *do not* specify `--peers` as we will manually define a replicator after startup via the `rpc` client command.

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
defradb client rpc addreplicator "Article" /ip4/0.0.0.0/tcp/9172/p2p/<peerID_of_nodeB>
```

As we add or update documents in the "Article" collection on *nodeA*, they will be actively pushed to *nodeB*. Note that changes to *nodeB* will still be passively published back to *nodeA*, via pubsub.

## Securing the HTTP API with TLS

By default, DefraDB will expose the HTTP API at `http://localhost:9181/api/v0`. It's also possible to configure the API to use TLS with self signed certificates or Let's Encrypt.

To start defradb with self signed certificates placed under `~/.defradb/certs/` with `server.key`
being the public key and `server.crt` being the private key, just do:
```shell
defradb start --tls
```

The keys can be generated with your generator of choice or with `make tls-certs`.

Since the keys should be stored within the DefraDB data and configuration directory, the recommended key generation command is `make tls-certs path="~/.defradb/certs"`.

If not saved under `~/.defradb/certs` then the public (`pubkeypath`) and private (`privkeypaths`) key paths need to be
explicitly defined inaddition to the `--tls` flag or `tls` set to `true` in the config.

Then to start the server with TLS, using your generated keys in custom path:
```shell
defradb start --tls --pubkeypath ~/path-to-pubkey.key --privkeypath ~/path-to-privkey.crt

```

Note the following example can sometimes not properly expand `~` and can cause problems:
```shell
defradb start --tls --pubkeypath="~/path-to-pubkey.key" --privkeypath="~/path-to-privkey.crt"
```

DefraDB also comes with automatic HTTPS for deployments on the public web. To enable HTTPS,
 deploy DefraDB to a server with both port 80 and port 443 open. With your domain's DNS A record
 pointed to the IP of your server, you can run the database using the following command:
```shell
sudo defradb start --tls --url=your-domain.net --email=email@example.com
```
Note: `sudo` is needed above for the redirection server (to bind port 80).

A valid email address is necessary for the creation of the certificate, and is important to get notifications from the Certificate Authority - in case the certificate is about to expire, etc.


## Licensing

DefraDB's code is released under the [Business Source License (BSL)](licenses/BSL.txt). It grants you the right to copy, modify, create derivative works, redistribute, and make non-production use of it. For additional uses, such as deploying in production on a private network, please contact license@source.network for a licensing agreement. Each dated version of the license turns into the more permissive Apache License v2.0 after four years. Please read the complete license before usage.



## Contributors

- John-Alan Simmons ([@jsimnz](https://github.com/jsimnz))
- Andrew Sisley ([@AndrewSisley](https://github.com/AndrewSisley))
- Shahzad Lone ([@shahzadlone](https://github.com/shahzadlone))
- Orpheus Lummis ([@orpheuslummis](https://github.com/orpheuslummis))
- Fred Carle ([@fredcarle](https://github.com/fredcarle))
- Islam Aliev ([@islamaliev](https://github.com/islamaliev))

You are invited to contribute to DefraDB. Follow the [Contributing guide](./CONTRIBUTING.md) to get started.
