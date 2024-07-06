---
sidebar_label: Akash Deployment Guide
sidebar_position: 60
---
# Deploy DefraDB on Akash

## Overview

This guide will walk you through the required steps to deploy DefraDB on Akash.

## Prerequisites

Before you get started you will need an Akash account with at least 5 AKT. If don't have an Akash account you can create one by installing [Keplr](https://www.keplr.app/).

## Deploy

![Cloudmos console](/img/akash/deploy.png "Cloudmos console")

Deploying on Akash can be done through the [Cloudmos console](https://deploy.cloudmos.io/new-deployment). Click on the "Empty" deployment type and copy the config below into the editor.

```yaml
---
version: "2.0"

services:
  defradb:
    image: sourcenetwork/defradb:develop
    args:
      - start
      - --url=0.0.0.0:9181
    expose:
      - port: 9171
        as: 9171
        to:
          - global: true
      - port: 9181
        as: 80
        to:
          - global: true

profiles:
  compute:
    defradb:
      resources:
        cpu:
          units: 1.0
        memory:
          size: 1Gi
        storage:
          size: 1Gi
  placement:
    akash:
      attributes:
        host: akash
      signedBy:
        anyOf:
          - "akash1365yvmc4s7awdyj3n2sav7xfx76adc6dnmlx63"
          - "akash18qa2a2ltfyvkyj0ggj3hkvuj6twzyumuaru9s4"
      pricing:
        defradb: 
          denom: uakt
          amount: 10000

deployment:
  defradb:
    akash:
      profile: defradb
      count: 1 
```

Next click the "Create Deployment" button. A pop-up will appear asking you to confirm the configuration transaction.

After confirming you will be prompted to select a provider. Select a provider with a price and location that makes sense for your use case.

A final pop-up will appear asking you to confirm the deployment transaction. If the deployment is successful you should now see deployment info similar to the image below.

## Deployment Info

![Cloudmos deployment](/img/akash/info.png "Cloudmos deployment")

To configure and interact with your DefraDB node, you will need the P2P and API addresses. They can be found at the labeled locations in the image above.

## P2P Replication

To replicate documents from a local DefraDB instance to your Akash deployment you will need to create a shared schema on both nodes.

Run the commands below to create the shared schema. 

First on the local node:

```bash
defradb client schema add '
    type User {
        name: String
        age:  Int
    }
'
```

Then on the Akash node:

```bash
defradb client schema add --url <api_address> '
    type User {
        name: String
        age:  Int
    }
'
```

> The API address can be found in the [deployment info](#deployment-info).

Next you will need the peer ID of the Akash node. Run the command below to view the node's peer info. 

```bash
defradb client p2p info --url <api_address>
```

If the command is successful, you should see output similar to the text below.

```json
{
  "ID": "12D3KooWQr7voGBQPTVQrsk76k7sYWRwsAdHRbRjXW39akYomLP3",
  "Addrs": [
    "/ip4/0.0.0.0/tcp/9171"
  ]
}
```

> The address here is the node's p2p bind address. The public p2p address can be found in the [deployment info](#deployment-info).

Setup the replicator from your local node to the Akash node by running the command below.

```bash
defradb client p2p replicator set --collection User '{
    "ID": "12D3KooWQr7voGBQPTVQrsk76k7sYWRwsAdHRbRjXW39akYomLP3", 
    "Addrs": [
        "/dns/<p2p_address_host>/<p2p_address_port>"
    ]
}'
```

> The p2p host and port can be found in the [deployment info](#deployment-info). For example: if your p2p address is http://provider.bdl.computer:32582/ the host would be provider.bdl.computer and the port would be 32582.

The local node should now be replicating all User documents to the Akash node.