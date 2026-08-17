![Test Coverage Workflow](https://github.com/sourcenetwork/defradb/actions/workflows/test-coverage.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/sourcenetwork/defradb)](https://goreportcard.com/report/github.com/sourcenetwork/defradb)
[![codecov](https://codecov.io/gh/sourcenetwork/defradb/branch/develop/graph/badge.svg?token=RHAORX13PA)](https://codecov.io/gh/sourcenetwork/defradb)
[![Discord](https://img.shields.io/discord/1374835078781468722.svg?color=768AD4&label=discord&logo=https%3A%2F%2Fdiscordapp.com%2Fassets%2F8c9701b98ad4372b58f13fd9f65f966e.svg)](https://source.network/discord)
[![X URL](https://img.shields.io/twitter/follow/edgeofsource.svg?label=&style=social)](https://x.com/edgeofsource)

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/DefraDB_White.svg">
    <img height="120px" width="374px" alt="DefraDB" src="docs/DefraDB_Full.svg">
  </picture>
</p>

DefraDB is a zero-trust database that prioritizes data verifiability, privacy, and information security. Its data model, powered by the convergence of [MerkleCRDTs](https://arxiv.org/pdf/2004.00107.pdf) and the content-addressability of [IPLD](https://docs.ipld.io/), enables a multi-write-master architecture. It features [DQL](https://docs.source.network/defradb/references/query-specification/query-language-overview), a query language compatible with GraphQL but providing extra convenience. By leveraging peer-to-peer networking it can be deployed nimbly in novel topologies. Access control is determined by a relationship-based DSL, supporting document or field-level policies, secured by the SourceHub network. DefraDB is a core part of the [Source technologies](https://source.network/) that enable new paradigms of decentralized data and access-control management, user-centric apps, data trustworthiness, and much more.

Read the documentation on [docs.source.network](https://docs.source.network/).


## Table of Contents

<!--ts-->
   * [Install](#install)
   * [Build Requirements](#build-requirements)
      * [Prerequisites](#prerequisites)
      * [System Resources](#system-resources)
      * [Building on Resource-Constrained Systems](#building-on-resource-constrained-systems)
   * [Key Management](#key-management)
   * [Start](#start)
   * [Configuration](#configuration)
   * [External port binding](#external-port-binding)
   * [Add a collection](#add-a-collection)
   * [Add a document](#add-a-document)
   * [Query documents](#query-documents)
   * [Obtain document commits](#obtain-document-commits)
   * [DefraDB Query Language (DQL)](#defradb-query-language-dql)
   * [Peer-to-peer data synchronization](#peer-to-peer-data-synchronization)
   * [Securing the HTTP API with TLS](#securing-the-http-api-with-tls)
   * [Access Control System](#access-control-system)
   * [Supporting CORS](#supporting-cors)
   * [Backing up and restoring](#backing-up-and-restoring)
   * [Telemetry](#telemetry)
   * [Community](#community)
   * [Explorer](#explorer)
   * [Licensing](#licensing)
   * [Contributors](#contributors)
<!--te-->

DISCLAIMER: The software is provided "as is" and is not guaranteed to be stable, secure, or error-free. We encourage you to experiment with DefraDB and provide feedback, and when you plan to deploy it to production, please thoroughly test your integrations.

## Install

Install `defradb` by [downloading an executable](https://github.com/sourcenetwork/defradb/releases) or building it locally using the [Go toolchain](https://golang.org/):

```sh
git clone https://github.com/sourcenetwork/defradb.git
cd defradb
make install
```

In the following sections, we assume that `defradb` is included in your `PATH`. If you installed it with the Go toolchain, use:

```sh
export PATH=$PATH:$(go env GOPATH)/bin
```

We recommend experimenting with queries using a native GraphQL client. GraphiQL is a popular option - [download and install it](https://altairgraphql.dev/#download).

## Build Requirements

Building DefraDB from source requires significant system resources. If you encounter out-of-memory errors or build failures, review the requirements below.

### Prerequisites

- [Go](https://golang.org/) 1.24 or later
- [Rust toolchain](https://www.rust-lang.org/tools/install) (for WASM lens compilation, if running tests)
- Git

### System Resources

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| RAM | 2 GB | 4+ GB |
| Disk Space | 3 GB | 5+ GB |

The Go compiler requires substantial memory during compilation. Builds with less than 2 GB of available RAM will likely fail with out-of-memory errors.

### Building on Resource-Constrained Systems

If you're building on a system with limited RAM (e.g., a small VM or container), you may encounter build failures. Common issues and solutions:

**Out of Memory (OOM) errors:**
- Ensure at least 2 GB of RAM is available
- Add swap space if physical RAM is limited
- Use `-p 1` to limit compiler parallelism: `go build -p 1 ./cmd/defradb`

**`/tmp` running out of space:**
On systems where `/tmp` is a small tmpfs (RAM-backed filesystem), the Go compiler may exhaust available space. Redirect Go's temp directories to a location with more space:
```sh
export GOTMPDIR=/path/with/space
export GOCACHE=/path/with/space/go-cache
export GOMODCACHE=/path/with/space/go-mod
```

**Reducing memory usage:**
For extremely constrained environments, disable optimizations (produces slower binary):
```sh
go build -p 1 -gcflags="all=-N -l" ./cmd/defradb
```

## Key Management

DefraDB has a built in keyring that can be used to store private keys securely.

The following keys are loaded from the keyring on start:

- `peer-key` Ed25519 private key (required)
- `encryption-key` AES-128, AES-192, or AES-256 key (optional)
- `node-identity-key` Secp256k1 private key (optional). This key is used for node's identity.

A secret to unlock the keyring is required on start and must be provided via the `DEFRA_KEYRING_SECRET` environment variable. If a `.env` file is available in the working directory, the secret can be stored there or via a file at a path defined by the `--secret-file` flag.

The keys will be randomly generated on the initial start of the node if they are not found.

Alternatively, to randomly generate the required keys, run the following command:

Node identity is an identity assigned to the node. It is used to exchange encryption keys with other nodes. 

```
defradb keyring new
```

To import externally generated keys, run the following command:

```
defradb keyring add <name> <private-key-hex>
```

To learn more about the available options:

```
defradb keyring --help
```

## Start

Start a node by executing `defradb start`. Keep the node running while going through the following examples.

Verify the local connection to the node works by executing `defradb client collection describe` in another terminal.

## Configuration

In this document, we use the default configuration, which has the following behavior:

- `~/.defradb/` is DefraDB's configuration and data directory
- `client` command interacts with the locally running node
- The GraphQL endpoint is provided at <http://localhost:9181/api/graphql> and a versioned API endpoint is provided at <http://localhost:9181/api/v1/graphql>

The GraphQL endpoint can be used with a GraphQL client (e.g., Altair) to conveniently perform requests (`query`, `mutation`) and obtain schema introspection.

Read more about the configuration [here](./docs/config.md).

## External port binding

By default the HTTP API and P2P network will use localhost. If you want to expose the ports externally you need to specify the addresses in the config or command line parameters.

```
defradb start --p2paddr /ip4/0.0.0.0/tcp/9171 --url 0.0.0.0:9181
```

## Add a collection

Collections are used to structure documents using a type system.

In the following examples, we'll be using a simple `User` type.

Add it to the database with the following command. By doing so, DefraDB generates the typed GraphQL endpoints for querying, mutation, and introspection.

```shell
defradb client collection add '
  type User {
    name: String
    age: Int
    verified: Boolean
    points: Float
  }
'
```

Find more examples of type definitions in the [examples/collection/](examples/collection/) folder.

## Add a document

Submit a `mutation` request to add a document of the `User` type:

```shell
defradb client query '
  mutation {
      add_User(input: {age: 31, verified: true, points: 90, name: "Bob"}) {
          _docID
      }
  }
'
```

Expected response:

```json
{
  "data": {
    "add_User": [
      {
        "_docID": "bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab",
      }
    ]
  }
}
```

`_docID` is the document's unique identifier determined by its collection and initial data.

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
    User(filter: {points: {_geq: 50}}) {
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

DefraDB's data model is based on [MerkleCRDTs](https://arxiv.org/pdf/2004.00107.pdf). Each document has a graph of all of its updates, similar to Git. The updates are called `commit`s and are identified by `cid`, a content identifier. Each references its parents by their `cid`s.

To get the most recent commit in the MerkleDAG for the document identified as `bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab`:

```shell
defradb client query '
  query {
    _commits(docID: "bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab") {
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
    "_commits": [
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
    _commits(cid: "bafybeifhtfs6vgu7cwbhkojneh7gghwwinh5xzmf7nqkqqdebw5rqino7u") {
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

Read its documentation at [docs.source.network](https://docs.source.network/defradb/references/query-specification/query-language-overview) to discover its filtering, ordering, limiting, relationships, variables, aggregate functions, and other useful features.

## Peer-to-peer data synchronization

DefraDB leverages peer-to-peer networking for data exchange, synchronization, and replication of documents and commits.

When starting a node for the first time, a key pair is generated and stored in its "root directory" (`~/.defradb/` by default).

Each node has a unique `PeerID` generated from its public key. This ID allows other nodes to connect to it.

To view your node's peer info:

```shell
defradb client p2p info
```

There are two types of peer-to-peer relationships supported: **pubsub** peering and **replicator** peering.

Pubsub peering *passively* synchronizes data between nodes by broadcasting *Document Commit* updates to the topic of the commit's document key. Nodes need to be listening on the pubsub channel to receive updates. This is for when two nodes *already* have shared a document and want to keep them in sync.

Replicator peering *actively* pushes changes from a specific collection *to* a target peer.

<details>
<summary>Pubsub example</summary>

Pubsub peers can be specified on the command line using the `--peers` flag, which accepts a comma-separated list of peer [multiaddresses](https://docs.libp2p.io/concepts/addressing/). For example, a node at IP `192.168.1.12` listening on 9000 with PeerID `12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B` would be referred to using the multiaddress `/ip4/192.168.1.12/tcp/9000/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B`.

Let's go through an example of two nodes (*nodeA* and *nodeB*) connecting with each other over pubsub, on the same machine.

Start *nodeA* with a default configuration:

```shell
defradb start
```

Obtain the node's peer info:

```shell
defradb client p2p info
```

In this example, we use `12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B`, but locally it will be different.

For *nodeB*, we provide the following configuration:

```shell
defradb start --rootdir ~/.defradb-nodeB --url localhost:9182 --p2paddr /ip4/127.0.0.1/tcp/9172 --peers /ip4/127.0.0.1/tcp/9171/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B
```

About the flags:

- `--rootdir` specifies the root dir (config and data) to use
- `--url` is the address to listen on for the client HTTP and GraphQL API
- `--p2paddr` is a comma-separated list of multiaddresses to listen on for p2p networking
- `--peers` is a comma-separated list of peer multiaddresses

This starts two nodes and connects them via pubsub networking.
</details>

<details>
<summary>Subscription example</summary>

It is possible to subscribe to updates on a given collection by using its ID as the pubsub topic. The ID of a collection is found as the field `collectionID` in one of its documents. Here we use the collection ID of the `User` type we created above. After setting up 2 nodes as shown in the [Pubsub example](#pubsub-example) section, we can subscribe to collections updates on *nodeA* from *nodeB* by using the following command:

```shell
defradb client p2p collection add --url localhost:9182 bafkreibpnvkvjqvg4skzlijka5xe63zeu74ivcjwd76q7yi65jdhwqhske
```

Multiple collection IDs can be added at once.

```shell
defradb client p2p collection add --url localhost:9182 <collection1ID>,<collection2ID>,<collection3ID>
```
</details>

<details>
<summary>Replicator example</summary>

Replicator peering is targeted: it allows a node to actively send updates to another node. Let's go through an example of *nodeA* actively replicating to *nodeB*:

Start *nodeA*:

```shell
defradb start
```

In another terminal, add this example collection to it:

```shell
defradb client collection add '
  type Article {
    content: String
    published: Boolean
  }
'
```

Start (or continue running from above) *nodeB*, that will be receiving updates:

```shell
defradb start --rootdir ~/.defradb-nodeB --url localhost:9182 --p2paddr /ip4/0.0.0.0/tcp/9172
```

Here we *do not* specify `--peers` as we will manually define a replicator after startup via the `rpc` client command.

In another terminal, add the same collection to *nodeB*:

```shell
defradb client collection add --url localhost:9182 '
  type Article {
    content: String
    published: Boolean
  }
'
```

Then copy the peer info from *nodeB*:

```shell
defradb client p2p info --url localhost:9182
```

Set *nodeA* to actively replicate the Article collection to *nodeB*:

```shell
defradb client p2p replicator add -c Article <nodeB_peer_info_json>
```

As we add or update documents in the Article collection on *nodeA*, they will be actively pushed to *nodeB*. Note that changes to *nodeB* will still be passively published back to *nodeA*, via pubsub.
</details>

## Securing the HTTP API with TLS

By default, DefraDB exposes its HTTP API over plain HTTP at `http://localhost:9181/api`. It can instead serve the API over HTTPS using a TLS certificate.

DefraDB enables TLS automatically when both the certificate (`server.crt`) and the private key (`server.key`) are present in the `certs` directory inside the data and configuration directory (by default `~/.defradb/certs/`). With both files in place, just start the node:
```shell
defradb start
```
and the HTTP API is served over `https://localhost:9181` instead.

Enabling TLS requires both files: if only one of `server.crt` and `server.key` is present (or only one of the paths below is set), `defradb start` fails with an error rather than silently starting without TLS. With neither present, DefraDB simply starts over plain HTTP.

The certificate and key can be generated with your generator of choice, or with `make tls-certs`. Since they should live inside the DefraDB data and configuration directory, the recommended command is:
```shell
make tls-certs path="~/.defradb/certs"
```

To use a certificate and key stored elsewhere, set both paths explicitly (both are required). The certificate path is `pubkeypath` and the private key path is `privkeypath`; they can be passed as flags or set in the config file:
```shell
defradb start --pubkeypath ~/path-to/server.crt --privkeypath ~/path-to/server.key
```

Because the certificates are self-signed, HTTPS clients must be configured to trust them (for example, `curl -k`). Note that the bundled `defradb` CLI does not yet connect to a TLS-enabled node.

## Access Control System
Learn more about the [Document Access Control](https://docs.source.network/defradb/security/document-access-control/) system.

### Using SourceHub for Document ACP

By default, Document ACP is local to the node. Setting `acp.document.type` to `source-hub` stores policies and relationships on a [SourceHub](https://github.com/sourcenetwork/sourcehub) chain instead, which is what makes them verifiable across nodes.

Four parameters are required on the node when that type is selected, and one more on the client:

| Parameter | Side | Description |
| --- | --- | --- |
| `acp.document.sourceHub.ChainID` | node | ID of the chain to write ACP data to |
| `acp.document.sourceHub.GRPCAddress` | node | address of the SourceHub gRPC server |
| `acp.document.sourceHub.CometRPCAddress` | node | address of the SourceHub Comet RPC server |
| `acp.document.sourceHub.KeyName` | node | keyring entry holding the secp256k1 key that signs and pays for SourceHub transactions |
| `acp.document.sourceHub.address` | client | SourceHub address allowed to act on the client's behalf |

The signing key is a regular keyring entry, so it is added the same way as any other key:

```shell
defradb keyring add sourcehub-key <secp256k1-private-key-hex>
```

#### Running a SourceHub node locally

For local development, the SourceHub image can run an isolated chain of its own with `STANDALONE=1`, which is also how the integration tests bring it up:

```shell
docker run --rm \
  -e STANDALONE=1 \
  -p 26657:26657 -p 9090:9090 \
  ghcr.io/sourcenetwork/sourcehub:dev
```

It exposes gRPC on `9090` and Comet RPC on `26657`, and the chain is `sourcehub-dev`. The standalone image also creates a funded `faucet` account, whose mnemonic is printed in the container logs.

The account behind `KeyName` pays for the SourceHub transactions the node creates, and the faucet does not fund it automatically. Either import the faucet mnemonic as the node's key, or send funds to the address derived from the key already in the keyring:

```shell
docker exec <container> sourcehubd tx bank send \
  faucet <address of the KeyName key> 1000000uopen \
  --keyring-backend test --chain-id sourcehub-dev --yes
```

A node is pointed at it through the config file at `~/.defradb/config.yaml`:

```yaml
acp:
  document:
    type: source-hub
    sourceHub:
      ChainID: sourcehub-dev
      GRPCAddress: 127.0.0.1:9090
      CometRPCAddress: 127.0.0.1:26657
      KeyName: sourcehub-key
```

Only `acp.document.type` and `acp.document.sourceHub.address` have CLI flags (`--document-acp-type` and `--source-hub-address`); the rest are set in the config file or through the environment, where `DEFRA_` is prefixed and dots become underscores:

```shell
export DEFRA_ACP_DOCUMENT_TYPE=source-hub
export DEFRA_ACP_DOCUMENT_SOURCEHUB_CHAINID=sourcehub-dev
export DEFRA_ACP_DOCUMENT_SOURCEHUB_GRPCADDRESS=127.0.0.1:9090
export DEFRA_ACP_DOCUMENT_SOURCEHUB_COMETRPCADDRESS=127.0.0.1:26657
export DEFRA_ACP_DOCUMENT_SOURCEHUB_KEYNAME=sourcehub-key
defradb start
```

The SourceHub version DefraDB is built against is pinned in `go.mod`, currently `github.com/sourcenetwork/sourcehub v0.4.1-0.20260128164915-1bce44032618`; use a node built from that revision when running against something other than the `dev` image.

## Supporting CORS

When accessing DefraDB through a frontend interface, you may be confronted with a CORS error. That is because, by default, DefraDB will not have any allowed origins set. To specify which origins should be allowed to access your DefraDB endpoint, you can specify them when starting the database:
```shell
defradb start --allowed-origins=https://yourdomain.com
```

If running a frontend app locally on localhost, allowed origins must be set with the port of the app:
```shell
defradb start --allowed-origins=http://localhost:3000
```

The catch-all `*` is also a valid origin. 

## Backing up and restoring

It is currently not possible to do a full backup of DefraDB that includes the history of changes through the Merkle DAG. However, DefraDB currently supports a simple backup of the current data state in JSON format that can be used to seed a database or help with transitioning from one DefraDB version to another.

To backup the data, run the following command:
```shell
defradb client backup export path/to/backup.json
```

To pretty print the JSON content when exporting, run the following command:
```shell
defradb client backup export --pretty path/to/backup.json
```

To restore the data, run the following command:
```shell
defradb client backup import path/to/backup.json
```

## Telemetry

DefraDB has no telemetry reporting by default. To enable OpenTelemetry in DefraDB you must build with the `telemetry` tag set. To configure the HTTP exporters use the environment variables in the links below.

[Metric exporter documentation](https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp)

[Trace exporter documentation](https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp)

## Community

Discuss on [Discord](https://source.network/discord) or [Github Discussions](https://github.com/sourcenetwork/defradb/discussions). The Source project is on [X](https://x.com/edgeofsource).

## Explorer

Instructions for the explorer can be found [here](./explorer/README.md).

## Licensing

DefraDB's code is released under the [Business Source License (BSL)](licenses/BSL.txt). It grants you the right to copy, modify, create derivative works, redistribute, and make non-production use of it. For additional uses, such as deploying in production on a private network, please contact license@source.network for a licensing agreement. Each dated version of the license turns into the more permissive Apache License v2.0 after four years. Please read the complete license before usage.

## Contributors

- John-Alan Simmons ([@jsimnz](https://github.com/jsimnz))
- Andrew Sisley ([@AndrewSisley](https://github.com/AndrewSisley))
- Shahzad Lone ([@shahzadlone](https://github.com/shahzadlone))
- Orpheus Lummis ([@orpheuslummis](https://github.com/orpheuslummis))
- Fred Carle ([@fredcarle](https://github.com/fredcarle))
- Islam Aliev ([@islamaliev](https://github.com/islamaliev))
- Keenan Nemetz ([@nasdf](https://github.com/nasdf))
- Ivan Vercenco ([@iverc](https://github.com/iverc))
- Chris Quigley ([@ChrisBQu](https://github.com/ChrisBQu))
- Jack Zampolin ([@jackzampolin](https://github.com/jackzampolin))

You are invited to contribute to DefraDB. Follow the [Contributing guide](./CONTRIBUTING.md) to get started.
