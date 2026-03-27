![Test Coverage Workflow](https://github.com/sourcenetwork/defradb/actions/workflows/test-coverage.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/sourcenetwork/defradb)](https://goreportcard.com/report/github.com/sourcenetwork/defradb)
[![codecov](https://codecov.io/gh/sourcenetwork/defradb/branch/develop/graph/badge.svg?token=RHAORX13PA)](https://codecov.io/gh/sourcenetwork/defradb)
[![Discord](https://img.shields.io/discord/427944769851752448.svg?color=768AD4&label=discord&logo=https%3A%2F%2Fdiscordapp.com%2Fassets%2F8c9701b98ad4372b58f13fd9f65f966e.svg)](https://discord.gg/w7jYQVJ)
[![Twitter Follow](https://img.shields.io/twitter/follow/sourcenetwrk.svg?label=&style=social)](https://twitter.com/sourcenetwrk)

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/DefraDB_White.svg">
    <img height="120px" width="374px" alt="DefraDB" src="docs/DefraDB_Full.svg">
  </picture>
</p>

DefraDB is a user-centric database that prioritizes data ownership, personal privacy, and information security. Its data model, powered by the convergence of [MerkleCRDTs](https://arxiv.org/pdf/2004.00107.pdf) and the content-addressability of [IPLD](https://docs.ipld.io/), enables a multi-write-master architecture. It features [DQL](https://docs.source.network/defradb/references/query-specification/query-language-overview), a query language compatible with GraphQL but providing extra convenience. By leveraging peer-to-peer networking it can be deployed nimbly in novel topologies. Access control is determined by a relationship-based DSL, supporting document or field-level policies, secured by the SourceHub network. DefraDB is a core part of the [Source technologies](https://source.network/) that enable new paradigms of decentralized data and access-control management, user-centric apps, data trustworthiness, and much more.

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
   * [Contributing](#contributing)
   * [License](#license)

## Install

To install the latest version of DefraDB, ensure you have [Go](https://go.dev/doc/install) installed (version 1.21 or later), then run:

```bash
go install github.com/sourcenetwork/defradb/cmd/defradb@latest
```

Alternatively, you can clone the repository and build from source:

```bash
git clone https://github.com/sourcenetwork/defradb.git
cd defradb
make build
```

## Build Requirements

### Prerequisites

DefraDB requires the following tools to be installed:
- Go (1.21+)
- GCC / C Compiler (for SQLite/Badger dependencies)
- Make

### System Resources

Building DefraDB requires a minimum of 2GB of RAM. If you are building on a machine with limited memory, the build might fail during the linking stage.

### Building on Resource-Constrained Systems

If you encounter memory issues during compilation, try limiting the number of parallel jobs:
```bash
GOMAXPROCS=1 go build ./cmd/defradb
```

## Key Management
... (rest of the sections)

## Contributing

We welcome contributions! Please check out our [Contributing Guidelines](CONTRIBUTING.md) to get started. You can also join our community on [Discord](https://discord.gg/w7jYQVJ) to discuss features and get help.

## License

DefraDB is licensed under the [Apache License, Version 2.0](LICENSE).