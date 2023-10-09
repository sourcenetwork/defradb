![Tests Workflow](https://github.com/sourcenetwork/defradb/actions/workflows/test-and-upload-coverage.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/sourcenetwork/defradb)](https://goreportcard.com/report/github.com/sourcenetwork/defradb)
[![codecov](https://codecov.io/gh/sourcenetwork/defradb/branch/develop/graph/badge.svg?token=RHAORX13PA)](https://codecov.io/gh/sourcenetwork/defradb)
[![Discord](https://img.shields.io/discord/427944769851752448.svg?color=768AD4&label=discord&logo=https%3A%2F%2Fdiscordapp.com%2Fassets%2F8c9701b98ad4372b58f13fd9f65f966e.svg)](https://discord.source.network/)
[![Twitter Follow](https://img.shields.io/twitter/follow/sourcenetwrk.svg?label=&style=social)](https://twitter.com/sourcenetwrk)

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/DefraDB_White.svg">
    <img height="120px" width="374px" alt="DefraDB" src="docs/DefraDB_Full.svg">
  </picture>
</p>

DefraDB is a user-centric database that prioritizes data ownership, personal privacy, and information security. Its data model, powered by the convergence of [MerkleCRDTs](https://arxiv.org/pdf/2004.00107.pdf) and the content-addressability of [IPLD](https://docs.ipld.io/), enables a multi-write-master architecture. It features [DQL](https://docs.source.network/references/query-specification/query-language-overview), a query language compatible with GraphQL but providing extra convenience. By leveraging peer-to-peer networking it can be deployed nimbly in novel topologies. Access control is determined by a relationship-based DSL, supporting document or field-level policies, secured by the SourceHub network. DefraDB is a core part of the [Source technologies](https://source.network/) that enable new paradigms of decentralized data and access-control management, user-centric apps, data trustworthiness, and much more.

## Getting Started

Follow the [Quick Start](docs/quick-start.md) guide to get started with DefraDB. 

Read more documentation on [docs.source.network](https://docs.source.network/).

## Community

Discuss on [Discord](https://discord.source.network/) or [Github Discussions](https://github.com/sourcenetwork/defradb/discussions). The Source project is on [Twitter](https://twitter.com/sourcenetwrk).

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
