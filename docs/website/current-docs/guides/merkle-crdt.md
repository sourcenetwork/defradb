---
sidebar_label: Merkle CRDT Guide
sidebar_position: 30
---
# A Guide to Merkle CRDTs in DefraDB

## Overview
Merkle CRDTs are a type of Conflict-free Replicated Data Type (CRDT). They are designed to update or modify independent sets of data without any human intervention, ensuring that updates made by multiple actors are merged without conflicts. The goal of Merkle CRDT is to perform deterministic, automatic data merging and synchronization without any inconsistencies. CRDTs were first formalized in 2011 and have become a useful tool in distributed computing. Merkle CRDTs are a new kind of CRDT that allows data to be merged without conflicts, ensuring that data is deterministically synchronized across multiple actors. This can be useful in a variety of distributed computing applications where data needs to be updated and merged in a consistent and conflict-free manner.

## Background on Regular CRDTs 
Conflict-free Replicated Data Types (CRDTs) are a useful tool in local and offline-first applications. They allow multiple actors or peers to collaborate and update the state of a data structure without worrying about synchronizing that state. CRDTs come in many different forms and can be applied to a variety of data types, such as simple registers, counters, sets, lists, and maps. The key feature of CRDTs is their ability to merge data deterministically, ensuring that all actors eventually reach the same state.

To achieve this, CRDTs rely on the concept of causality or ordering of events. This determines how the merge algorithm works and ensures that if all events or updates are applied to a data type, the resulting state will be the same for all actors. In distributed systems, however, the concept of time and causality can be more complex than it appears. This is because it is often difficult to determine the relative order of events occurring on different computers in a network. As a result, CRDTs often rely on some sort of clock or a different mechanism for tracking the relative order of events.

## Need for CRDTs

It can be difficult to determine the relative order of events occurring on different computers in a network, which is why CRDTs can enable the user to ensure data can be merged without conflicts. For example, consider a situation where two actors, A and B, are making updates to the same data at the same time. If actor A stamps their update with a system time of 2:39:56 PM EST on September 6, 2022, and actor B stamps their update with a system time of 2:40:00 PM, it would look like actor B's update occurred after actor A's. However, system times are not always reliable because they can be easily changed by actors, leading to inconsistencies in the relative order of events. To solve this problem, distributed systems use alternative clocks such as logical clocks or vector clocks to track the causality of events.


To track the relative causality of events, CRDTs often rely on clocks such as logical clocks or vector clocks. However, these clocks have limitations when used in high-churn networks with a large number of peers. For example, in a peer-to-peer network with a high rate of churn, logical and vector clocks require additional metadata for each peer that an actor interacts with. This metadata must be constantly maintained for each peer, which can be inefficient if the number of peers is unbounded. Additionally, in high churn environments, the amount of metadata grows linearly with the churn rate, making it infeasible to use these clocks in certain situations. Therefore, existing CRDT clock implementations may not be sufficient for use in high churn networks with an unbounded number of peers.

## Formalization of Merkle CRDT

Merkle CRDTs are a type of CRDT that combines traditional CRDTs with a new approach to CRDT clocks called a Merkle clock. This clock allows us to solve the issue of maintaining a constant amount of metadata per peer in a high churn network. Instead of tracking this metadata, we can use the inherent causality of Merkle DAGs (Directed Acyclic Graphs). In these graphs, each node is identified using its content identifiable data (CID) and is embedded in another node. The edges in these graphs are directed, meaning one node points to another, forming a DAG structure. If a node points to another node, the CID of the first node is embedded in the value of the second. The inherent nature of Merkle graphs is the embedded relation of hashing or CIDs from one node to another, providing us with useful properties.


To create a Merkle CRDT, we take an existing Merkle clock and embed any CRDT that satisfies the requirements. A CRDT is made up of three components: the data type, the CRDT type (operation-based or state-based), and the semantic type. For our specific implementation, we use delta state based CRDTs with different data types and semantic types for different applications. The formal structure of a CRDT is simple - it consists of a Merkle CRDT outer box containing two inner boxes, a Merkle clock and a regular CRDT.



## Merkle Clock

Merkle clocks are a type of clock used in distributed systems to solve the issue of tracking metadata for each peer that an actor interacts with. They are based on Merkle DAGs that function like hash chains, similar to a blockchain. These graphs are made up of nodes and edges, where the edges are directed, meaning that one node points to another. The head of a Merkle DAG is the most recent node added to the graph, and the entire graph can be referred to by the CID of the head node. The size of the CID hash does not grow with the number of nodes in the graph, making it a useful tool for high churn networks with a large number of peers.

The Merkle clock is created by adding an additional metadata field to each node of the Merkle DAG, called the height value, which acts as an incremental counter that increases with each new node added to the system. This allows the Merkle clock to provide a rough sense of causality, meaning that it can determine if one event happened before, at the same time, or after another event. The inherent causality of the Merkle DAG ensures that events are recorded in the correct order, making it a useful tool for tracking changes in a distributed system.

The embedding of CID into the parent node that produces the hash chain provides a causality guarantee that, for example a B is pointed to by node A, node C is pointed to by node B and so on till the node Z, A had to exist before B, because the value of A is embedded inside B, and B could not exist before A, otherwise it would result in breaking the causality of time because the value of A is embedded inside the value of B which then gets embedded inside the value of C, which means that C has to come after B and so on, all the way till the user gets back to Z. And hence if the user has constructed a Merkle DAG correctly, then A has to happen before B, B has to happen before C, C has to happen before D, all the way until they get to Z. This inherent causality of time with respect to CIDs and Merkle DAG provides the user with a causality-adhering system.

## Delta State Semantics

There are two types of Delta State Semantics: Operation-Based CRDTs and State-Based CRDTs. Operation-Based CRDTs use the intent of an operation as the body or content of the message, while State-Based CRDTs use the resulting state as the body or content of the message. Both have their own advantages and disadvantages, and the appropriate choice depends on the specific use case. Operation-Based CRDTs express actions such as setting a value to 10 or incrementing a counter by 4 through the intent of the operation. State-Based CRDTs, on the other hand, include the resulting state in the message. For example, a message to set a value to 10 would include the value 10 as the body or content of the message.

Operation-Based CRDTs tend to be smaller because their messages only contain the operation being performed, while State-Based CRDTs are larger because their messages contain both the current state and the state being changed. It is important to consider the trade-offs between these two types of Delta State Semantics when choosing which one to use in a given situation.

Delta State Semantics is an optimization of the State-based CRDTs. While both Operation-based CRDTs and State-based CRDTs have their own pros and cons, Delta State CRDTs offer a hybrid approach that uses the state as the message content, but with the same size as an operation.

In a Delta State CRDT, the message body includes only the minimum amount, or "delta," necessary to transform the previous state to the target state. For example, if we have a set of nine fruit names, and we want to add a banana to the set, the Delta State CRDT would only include the delta, or the value "banana," rather than expressing the entire set of 10 fruit names as in traditional State-based CRDTs. This is like an operation because it has the size of only one action, but it expresses the difference in state between the previous and target rather than the intent of the action.


## Branching and Merging State


### Branching of Merkle CRDTs


Merkle CRDTs are based on the concept of a Merkle clock, which is in turn based on the idea of a Merkle DAG. The structure of a Merkle DAG allows it to branch and merge at any point, as long as it adheres to the requirement of being a DAG and does not create a recursive loop.


Branching in a Merkle CRDT system occurs when two peers make independent changes to a common ancestor node and then share those changes, resulting in two distinct states. Neither of these states is considered the correct or canonical version in a Merkle CRDT system. Instead, both are treated as their own local main copies. From these divergent states, further updates can be made, causing the divergence to increase. For example, if there are 10 nodes in common between the two states, one branch may have five new nodes while the other has six. These branches exist independently of each other, and changes can be made to each branch independently without the need for immediate synchronization. This makes CRDTs useful for local-first or offline-first applications that can operate without network connectivity. The structure of a Merkle DAG, on which a Merkle CRDT is based, naturally supports branching.

### Merging of Merkle CRDTs

Merging in a Merkle CRDT system involves bringing two divergent states back together into a single, canonical graph. This is done by adding a new head node, known as a merge node, to the history of the graph. The merge node has two or more previous parents, as opposed to the traditional single parent of most nodes. To merge these states, merge semantics must be applied to the new system. The Merkle clock provides two pieces of information that facilitate this process: the use of a CID for each parent and the ability to go back in time through both branches of the divergent state, parent by parent, before officially merging the state. Each type of CRDT defines its own merge semantics.


The process begins by finding a common ancestral node between the two divergent states. Each node in the system includes a height parameter, which is the number of nodes preceding it. This, along with the CID of the ancestral node, is provided to the embedded CRDT's merge system to facilitate the merging process. The Merkle CRDT coordinates the logistics of the Merkle DAG and passes information about the multiple parents of the merge node to the embedded CRDT's merge system, which is responsible for defining the merge semantics. As long as the CRDT and the Merkle DAG are functioning correctly, the resulting Merkle clock will also operate correctly.

