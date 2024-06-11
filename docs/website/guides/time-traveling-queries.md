---
sidebar_label: Time Traveling Queries Guide
sidebar_position: 40
---
# A Guide to Time Traveling Queries in DefraDB

## Overview
Time Traveling queries allow users to query previous states of documents within the query interface. Essentially, it returns data as it had appeared at a specific commit. This is a powerful tool as it allows users to inspect and verify arbitrary states and time regardless of the number of updates made or who made these updates if the user has the current state. Since a current state is always going to be based on some previous state and that previous state is going to be based on another previous state, hence time-traveling queries provide the ability to “go back in time” and look at previous states with minimal changes to the working of the query. A special quality of this query is that there is minimal distinction between a regular query run versus a time-traveling query since both apply almost the same logic to fetch the result of the query.

## Background

The Web2 stack has traditional databases, like Postgres or MySQL, that usually have the current state as the only state. Once a user makes an update, the previous state is overwritten. There is no way to retrieve it from the system, unless a snapshot is captured, which exists as an independent file in the backup. The only way to access previous states is by loading the backup onto the database and querying the previous state. Additionally, in traditional systems, this backup occurs only once every hour, once a day, or once a month. This results in a loss of the ability to introspect each update made in the database. Here, the time travel inquiry system provides an edge over the existing databases as the data model of this system is independent of the mechanism of creating snapshots or backups that a user would utilize as part of natural maintenance and administration. The data model of time-traveling queries is such that every update is a function of all the previous updates. Essentially, there is no difference between inspecting the state of a document at a present point in time versus a previous point since the previous state is a function of the document graph.

## Usage

A powerful feature of a time-traveling query is that very little work is required from the developer to turn a traditional non-time-traveling query into a time-traveling query. Each update a document goes through gets a version identifier known as a Content Identifier (CID). CIDs are a function of the data model and are used to build out the time travel queries. These CIDs can be used to refer to a version that contains some piece of data. Instead of using some sort of human-invented notion of semantic version labels like Version 1, or Version 3.1 alpha, it uses the hash of the data as the actual identifier. The user can take the entire state of a document and create a single constant-sized CID. Each update in the document produces a new version number for the document, including a new version number for its individual fields. The developer then only needs to submit a new time-traveling query using the doc key of the document that it wants to query backward through its state, just like in a regular query, only here the developer needs to add the 32-bit hexadecimal version identifier that is expressed as it’s CID in an additional argument and the query will fetch the specific update that was made in the document.

```graphql
# Here we fetch a User of the given dockey, in the state that it was at
# at the commit matching the given CID.
query {
  User (
    cid: "bafybeieqnthjlvr64aodivtvtwgqelpjjvkmceyz4aqerkk5h23kjoivmu",
    dockey: "bae-d4303725-7db9-53d2-b324-f3ee44020e52"
  ) {
    name
    age
  }
}
```

## How It Works

The mechanism behind time-traveling queries is based on the Merkel CRDT system and the data model of the documents discussed in the above sections. Each time a document is updated, a log of updates known as the Update graph is recorded. This graph consists of every update that the user makes in the document from the beginning till the Nth update. In addition to the document update graph, we also have an independent and individual update graph for each field of the document. The document update graph would capture the overall updates made in the document whereas the independent and individual update graphs would capture the changes made to a specific field of the document. The data model as discussed in the Usage section works in a similar fashion, where it keeps appending the updates of the document to its present state. So even if a user deletes any information in the document, this will be recorded as an update within the update graph. Hence, no information gets deleted from the document, as all updates are stored in the update graph. 

[Include link to CRDT doc here]

Since we now have this update graph of changes, the query also takes its mechanism from the inherent properties of the Delta State Merkel CRDTs. Under this, the actual content of the update added by the user in the document is known as the Delta Payload. This delta payload is the amount of information that a user wants to go from a previous state to the very next state, where the value of the next state is set by some other user. For example, suppose a team of developers is working on a document and one of them wants to change the name of the document, then in this case, the delta payload of the new update would be the name of the document set by that user. Hence, the time-traveling queries work on two core concepts, the appending update graph and the delta payload which contains information that is required to go from the previous state to the next state. With both of these, whenever a user submits a regular query, the query caches the present state of the document within the database. And we internally issue a time-traveling query for the current state, with the only upside being that the user can submit a non-time-traveling query faster since a cached version of the same is already stored in the database. Thus, using this cached version of the present state of the document, the user can apply a time-traveling query using the CID of the specific version they want to query in the document. The database will then set the CID provided by the user as the Target State, a state at which the query will stop and go back to the beginning of the document, known as the Genesis State. The query will then apply all its operations until it reaches back to the Target State. 

The main reason behind setting a Target state is because the Merkel CRDT is a single-direction graph, and it only points backward in time. But to apply all the updates of all the delta payloads from the genesis to the target state, we need the query to track the existence of the target state as the present state of the target version can be a function of multiple operations. We thus perform a two-step process where it starts from the target version, goes to the genesis state, and comes back to the version. And from this, we produce the current present or the actual external facing state, also known as the serialized state.

## Limitations

1. Relational Limitation: A user will not be able to apply a time-traveling query to a series of related documents relating to the document that they are applying the query. For example, a person has some books and a list of their respective authors. An author can have many books under their name, but one book can be associated with one author only. Now, if a user applies a time-traveling to a specific version at some point in time of a particular book, it will only be able to query the state of that book and not the related state of its correlated author. A regular query, on the other hand, can go in-depth and present its values or get the state of the book and its correlated author. However, with the time-traveling query, the user will not be able to query beyond the exact state to which the query is applied.

2. Performance Limitation: As discussed earlier, the present state is stored as a cached version in the database, and based on this cached version, the current present state is computed. Hence the performance of the query depends on two factors:

    a. The size of the update graph of the document

    b. The number of updates that are between the Genesis state and the Target state. 

The larger the number of updates that exist between the Genesis state and the Target state, the longer it is going to take for the query to go back to the genesis state, perform its operations and come back to the target state. And hence, the time taken by the query to provide results would increase in proportion to the number of updates that are present between these two states.


## Future Outlook

The future outlook for time-traveling queries focuses mainly on resolving the current limitations that this query faces. To navigate the relational limitation, the current data model being used the time-traveling query needs to be exposed to the underlying aspects of the Merkel CRDT data model. Here, taking the help of the example mentioned for this limitation in the previous section, the relationship between the author and their book can be expressed by using doc keys of two types. Therefore, book A, which has a particular doc key (doc key A) to represent it, and this book has its author B, will be represented by its different doc key (doc key B). So, whenever the user correlates that book A was published by author B, the user relates the doc keys in the relationship field of the update graph. 

For the performance limitations, snapshots can be the next step toward the elimination of the performance limitation. Currently, we keep a cached version of the present or the current state of the document. Once an update is made, this cached version will get replaced with the current version. Therefore, at any given point in time, there would only be a single cached version of the current state. However, developers can choose to trade this space to decrease the time taken for the query execution. This can be achieved by creating snapshots at various points in the update history. For a document undergoing millions of updates, snapshots can be taken at every 1000th update and a cached version of this snapshot can be created such that if we need to query the 2000th update, we just need to go back to the closest snapshot instead of having to go back to the Genesis state and then moving 2000 states to get to the Target state. For example, if we need to query the 1010th update, then we only need to execute 10 steps backward and 10 steps forward from the cached update, i.e., the 1000th update. Therefore, depending on the interval of the cache set by the user, for every 'x' number of updates, they would be required to execute 'x' number of steps after the closest cached version of the snapshot.

