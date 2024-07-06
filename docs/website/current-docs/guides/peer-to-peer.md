---
sidebar_label: Peer-to-Peer Guide
sidebar_position: 10
---
# A Guide to Peer-to-Peer Networking in DefraDB

## Overview

P2P networking is a way for devices to communicate and share data directly with each other without the need for a central server. In a P2P network, all devices, also known as peers, are equal and can both send and receive data. DefraDB is a database that uses P2P networking instead of the traditional client-server model.

One advantage of this is that it allows for the development of offline-first or local-first applications. These are apps that can still work even when there is no internet connection and can sync data between multiple devices without the need for a central server to facilitate the synchronization. This makes it possible for a peer-to-peer network and database like DefraDB to function in a trustless environment, where no one device is more important or trustworthy than any other. This aligns with the goals of a decentralized, private, and user-centric database.

P2P networking is the primary method of communication used in DefraDB, a decentralized database. The libp2p library was developed specifically for this purpose and forms the technological foundation of the database. In DefraDB, documents are replicated and combined into an update graph, similar to a version control client like Git or a hash chain or a hash graph. P2P networking allows nodes in DefraDB to communicate directly with each other, without the need for an intermediate node, making it easier to synchronize the updates within the update graph of a document.

Libp2p is a decentralized network framework that enables the development of P2P applications. It consists of a set of protocols, specifications, and libraries created by Protocol Labs for the IPFS project. As the network layer for IPFS, libp2p provides various features for P2P communication such as transport, security, peer routing, and content discovery.

Libp2p is modular, meaning it can be customized and integrated into different P2P projects and applications. It is designed to work with the IPLD (Inter Planetary Linked Data) data model, which is a suite of technologies for representing and navigating hash-linked data. IPLD allows for the unification of all data models that link data with hashes as instances of IPLD, making it a suitable choice for use with libp2p in P2P networking.

## Documents and Collections

The high-level distinction between a document is as follows:

* A document is a single record that contains multiple fields. These documents are bound by schema. For example, each row in an SQL table has multiple individual columns. These rows are analogous to documents with multiple individual fields.

* A collection refers to a collection of documents under a single schema. For example, a table from an SQL database comprising of rows and columns is analogous to collections.

## Need for P2P Networking in DefraDB

The DefraDB database requires peer-to-peer (P2P) networking to facilitate data synchronization between nodes. This is necessary because DefraDB can store documents and individual IPLD blocks on various nodes around the world, which may be used by a single application or multiple applications. P2P networking allows local instances of DefraDB, whether on a single device or in a web browser, to replicate information with other devices owned by the user or with trusted third parties. These third parties may serve as historical archival nodes or may be other users with whom the user is collaborating. For example, if a collaborative document powered by DefraDB is being shared with others, it should be transmitted over a P2P network to avoid the need for a trusted intermediary node. DefraDB offers two types of replication over the P2P network:

* Passive replication

* Active replication

## How it works

There are two, concrete types of data replication within DefraDB, i.e., active, and passive replication. Both these replication types serve different use cases and are implemented using different mechanics.

### Passive Replication

In DefraDB, passive replication is a type of data replication in which updates are automatically broadcast to the network and its peers without explicit coordination. This occurs over a global publish-subscrib network (PubSub), which is a way to broadcast updates on a specific topic and receive updates on that topic. 

This is called passive replication because it is similar to a "fire and forget" scenario. Passive replication is enabled for all nodes by default and all nodes will always publish to the larger PubSub network. Passive replication can be compared to the connectionless protocol UDP, while active replication can be compared to the connection-oriented protocol TCP.

### Active Replication

In active replication, data is replicated between nodes in a direct, point-to-point manner. This means that a specific node is chosen to constantly receive updates from the local node. In contrast, passive replication uses the Gossip protocol, which is a peer-to-peer communication mechanism in which nodes exchange state information about themselves and other nodes they know about. In the Gossip protocol, each node initiates a gossip round every second to exchange information with another random node, and the process is repeated until the whole system is synchronized. One difference between active and passive replication is that the Gossip protocol is a multi-hop protocol, meaning that there may be multiple connections between nodes in the network. Active replication, on the other hand, creates a direct connection between two nodes and ensures that updates are actively pushed to the other node, which then acknowledges receipt of the update to establish two-way communication.

Passive replication is a good choice for situations where you want your peers to be able to follow your updates without requiring much coordination from you. It is often used in collaborative environments where multiple people are working on a document and want to ensure that both peers are in sync with each other. On the other hand, active replication is better for situations where you have a specific peer you are collaborating with and want to ensure that all of your data is being replicated to an archival node. This is because active replication involves a direct, point-to-point connection between the two nodes, allowing for more efficient and reliable data replication.

## Implementation of Peer-to-Peer Networking in DefraDB

In the DefraDB software architecture, a PubSub system is used for peer-to-peer networking. In this system, publishers send messages without specifying specific receivers, and subscribers express interest in certain types of messages without knowing which publishers they come from. This allows for a more dynamic network topology and better scalability. In the DefraDB PubSub network, nodes can publish or subscribe to specific topics. When a node publishes a message in passive replication, it is broadcasted to all nodes in the network. These nodes then coordinate with each other, re-broadcast the message, and use a process called "gossiping" to spread the published information through multiple connections, or "hops." This is known as the Gossip protocol.

In passive replication, updates are broadcasted on a per-document level over the global PubSub network. Each document has its own topic, and nodes can subscribe to the topic corresponding to a specific document to receive updates passively. This is useful in environments where certain documents are in high demand or are being frequently updated, as the connections to these "hot documents" can be kept open to ensure they are kept up-to-date. However, if a document has not been accessed in a while, it is less important for it to be constantly updated and it is easy to resync these "cold documents" by submitting a query for the relevant updates. Passive replication and the PubSub system are therefore focused on individual documents.

One major difference between active and passive networks is that an active network can focus on both collections and individual documents, while a passive network is only focused on individual documents. Active networks operate over a direct, point-to-point connection and allow you to select an entire collection to replicate to another node. For example, if you have a collection of books and specify a target node for active replication, the entire collection will be replicated to that node, including any updates to individual books. However, it is also possible to replicate granularly by selecting specific books within the collection for replication. Passive networks, on the other hand, are only concerned with replicating individual documents.

```bash
$ defradb client rpc addreplicator "Books" /ip4/0.0.0.0/tcp/9172/p2p/<peerID_of_node_to_replicate_to>
```

## Concrete Features of P2P in DefraDB

### Passive Replication Features

The Defra Command Line Interface (CLI) allows you to modify the behavior of the peer-to-peer data network. When a DefraDB node starts up, it is assigned a libp2p host by default.

```bash
$ defradb start
...
2023-03-20T07:18:17.276-0400, INFO, defra.cli, Starting P2P node, {"P2P address": "/ip4/0.0.0.0/tcp/9171"}
2023-03-20T07:18:17.281-0400, INFO, defra.node, Created LibP2P host, {"PeerId": "12D3KooWEFCQ1iGMobsmNTPXb758kJkFc7XieQyGKpsuMxeDktz4", "Address": ["/ip4/0.0.0.0/tcp/9171"]}
```

This host has a Peer ID, which is a function of a secret private key generated when the node is started for the first time. The Peer ID is important to know as it may be relevant for different parts of the peer-to-peer networking system. The libp2p networking stack can be enabled or disabled.

```bash
$ defradb start --no-p2p
```

The passive networking system can also be enabled or disabled. By default, if the P2P network is online, the passive networking system is turned on.

```bash
$ defradb start --peers /ip4/0.0.0.0/tcp/9171/p2p/<peerID_of_node_to_replicate_to>
```

A node automatically listens on multiple addresses or ports when the P2P module is instantiated. These are referred to as the peer-to-peer address, which is expressed as a multi-address. A multi-address is a string that represents a network address and includes information about the transport protocol and addresses for multiple layers of the network stack.


```bash
/ip4/0.0.0.0/tcp/9171/p2p/<peerID_of_node_to_replicate_to>

scheme/ip_address/protocol/port/protocol/peer_id
```
The peer listens in on the p2p port 9171​ by default, which can be customized through the CLI or the configuration file.

```bash
$ defradb start --p2paddr /ip4/0.0.0.0/tcp/9172
```

The peer-to-peer address is the first of the addresses that the peer listens in on.

At the start of a node, flags can be specified to enable, disable, or switch the host that the peer is listening on. When a new node is started, every existing or new document goes through an LRU (Least Recently Used) cache to identify the most important, relevant, or frequently used documents over a specific period of time. Then, by default, the passive replication system automatically subscribes to and creates the corresponding document topics on the PubSub network.

When a node is started, it specifies a list of peers that it wants to stay connected to. The peer-to-peer node is self-organizing, meaning that if a node joins a new topic, it asks the larger network for other peers that are sharing information on that topic. This ensures that the node is always connected to some relevant nodes. A node also tries to find other relevant nodes, particularly when an individual topic is joined, subscribed to, or published.

### Active Replication Features

To use the active replication feature in DefraDB, you can submit an add replicator Remote Procedure Call (RPC) command through the client API. You will need to specify the multi-address and Peer ID of the peer that you want to include in the replicator set, as well as the name of the collection that you want to replicate to that peer. These steps handle the process of defining which peers you want to connect to, enabling or disabling the underlying subsystems, and sending additional RPC commands to add any necessary replicators.

```bash
$ defradb client rpc addreplicator "Books" /ip4/0.0.0.0/tcp/9172/p2p/<peerID_of_node_to_replicate_to>
```

## Benefits of the P2P System

One of the main benefits of the peer-to-peer (P2P) system is its robustness and ability to work even in the event of network failures. This allows developers to create local-first, offline-first applications. If a developer's node loses its internet connection, the P2P system will continue making changes and queue up updates. When the system is back online and reconnects to the network, it will automatically resolve the updates and resume publishing or replicating to the nodes specified by the developer. This means that the developer can rely on a trustless mechanism and does not need to rely on a central, trusted peer for data replication or repositories to save data. Instead, data is directly passed from the developer's node to any other collaborating node. This global P2P network allows developers to collaborate with anyone across the internet with no fundamental limitations. Additionally, since the P2P system is built on top of libp2p, developers have access to other useful features as well. These factors make it highly advantageous to work with a P2P network, especially from a local-first perspective.

In DefraDB, the peer-to-peer system has several benefits. It is easy to connect to a server in a data center because each server has its own individual IP address. However, in a home network, there is a single IP for the modem and multiple devices connected to it are protected by a NAT firewall, making it difficult for other nodes to connect directly. The libp2p framework offers two solutions to this problem:

Circuit Relays - This allow you to specify a third-party node that acts as an intermediary to resolve the NAT firewall issue. This works when you connect to the firewall/circuit relay node, which is a publicly accessible node, and another node connects to it as well. The third-party node acts as a conduit in this situation. This process requires trust in the third-party node to properly relay information, but it operates over encrypted transport layers, so the third-party node cannot use man-in-the-middle attacks to listen in on the data exchange. However, it does require the third-party node to be online and accessible.

NAT Hole Punching - This is a technique that allows nodes to connect directly to a device behind a NAT firewall. This ensures that a user can directly connect with another node and vice versa, without the need for a trusted intermediary within the peer-to-peer network.

## Current Limitations and Future Outlook

Here are some of the limitations of the P2P system:

One limitation of the peer-to-peer system is the potential scalability issue with having every document have its own independent topic. This can lead to overhead if a user has thousands or tens of thousands of documents in their node, or if an application developer has hundreds of thousands or millions of documents in their node. To address this issue, the team is exploring ways to create aggregate topics that can be scoped to subnets. These subnets can be group-specific or application-specific. Multiple hops are required between subnets. This means that if a user wants to synchronize and broadcast updates from their subnet to another subnet, they have to go from their subnet to the global net and back to the other subnet. The team is exploring ways to navigate this limitation through multi-hop mechanisms.

In a peer-to-peer network, when a user broadcasts an update, it is sent to other nodes on the network. However, if a node is offline or experiences some other issue, it may miss some updates. In DefraDB's passive replication mode, the most recent update is broadcasted through the network using a Merkle DAG (directed acyclic graph). The broadcasting node does not verify that the receiving node has received all previous updates, so it is the responsibility of the receiving node to ensure it has received all necessary updates. If a node misses a couple of updates and then receives a new update, it must synchronize all previous updates before considering the document up to date. This is because the internal data model of the document is based on all changes made over time, not just the most recent change. When broadcasting the most recent update, it is sent over the peer-to-peer PubSub network. However, if a node needs to go back in time through the Merkle DAG to get updates from previous broadcasts, it uses a different system called the Distributed Hash Table (DHT).

The scalability of Bitswap and the Distributed Hash Table (DHT) have been identified as limitations in the peer-to-peer (P2P) system. To address these issues, we are exploring the use of two new protocols.:

PubSub based query system - This that allows users to query and receive updates through the global PubSub network using query topics that are independent of document topics.

Graph Sync - This is a protocol developed by Protocol Labs, which has the potential to resolve issues with the Bitswap algorithm and DHT. These two approaches show promise in improving the scalability of the P2P system.

There are currently some limitations with the peer-to-peer system being used. One issue is that replicators, which are added to a node, do not persist through updates or restarts. This means that the user must re-add the replicators every time the node is restarted. However, this issue will be resolved in the next version of the system.

Currently, when a replicator is added to a node, it doesn't persist between node updates or node restarts. This means that every time there is a restart, the user must re-add these replicators. This is a minor oversight that the Source team plans to fix in a future release. In the meantime, they are also wWorking on a new protocol called Head Exchange to address issues with syncing the Merkel DAG when updates have been missed or concurrent, diverged updates have been made. The Head Exchange protocol aims to efficiently establish the most recent update seen by each node, determine if there are any divergent updates, and figure out the most efficient way to synchronize the nodes with the least amount of communication.

One issue with peer-to-peer local-first development is that it can be difficult for nodes to connect with each other when they are running on devices within the same home Wi-Fi network. This is due to a NAT firewall, which is a router that operates to protect private networks. A NAT firewall only allows internet traffic to pass through if it was requested by a device on the private network. It protects the identity of a network by not exposing internal IP addresses to the internet. This can make it difficult for other nodes to connect directly to a node running behind a NAT firewall.
