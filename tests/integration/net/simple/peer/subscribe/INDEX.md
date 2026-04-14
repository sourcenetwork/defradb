# Index: `tests/integration/net/simple/peer/subscribe`

## Overview

This directory contains integration tests for the P2P subscription APIs in DefraDB, covering both collection-level and document-level subscription granularities. The tests verify the full lifecycle of subscriptions — adding, removing, and listing them — as well as confirming that subscribed peers correctly receive (or stop receiving) synced data, and that invalid inputs produce appropriate errors without disturbing existing subscriptions.

## Subdirectories

| Directory | Summary |
|---|---|
| [`collection/`](collection/INDEX.md) | Tests the collection-level subscription API: subscribing to entire collections, removing subscriptions, listing them, and verifying that document sync is scoped to subscribed collections only. |
| [`document/`](document/INDEX.md) | Tests the document-level subscription API: subscribing to specific documents by ID, removing subscriptions, listing them, and verifying that update sync is scoped to subscribed documents only. |
