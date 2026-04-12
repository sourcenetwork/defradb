---
sidebar_label: Mutation Block
sidebar_position: 150
---
# Mutation Block

Mutations are the `write` side of the DefraDB Query Language. They rely on the query system to function properly. Updates, upserts and deletes, all require filtering and finding data before taking action. 

The data and payload format that mutations use is fundamental to maintaining the designed structure of the database. All mutation definitions are generated for each defined type in the Database. This is similar to the read query system.

Mutations are similar to SQL `INSERT INTO ...` or `UPDATE` statements. Much like the Query system, all mutations exist inside a `mutation { ... }` block. Several mutations can be run at the same time, independently of one another.

## Insert

Insert is used to create new documents from scratch. This involves many necessary steps to ensure all the data is structured properly and verifiable. From a developer's perspective, it's the easiest of all the mutations as it doesn't require any queries or document lookups before execution.

```graphql
type Book { ... }

mutation {
    add_Book(input: [BookMutationInputArg!]) [Book]
}
```

The above example displays the general structure of an insert mutation. You call the `add_TYPE` mutation, with the given input.

Although the generated schema exposes `input` as a list type, GraphQL input coercion also allows a single input object in practice, which is why many examples use `input: { ... }`.

### Input Object Type

All mutations use a typed input object to update the data.

The following is an example with a full type and input object:

```graphql 
type Book {
    title: String
    description: String
    rating: Float
}

mutation {
    add_Book(input: {
        title: "Painted House",
        description: "The story begins as Luke Chandler ...",
        rating: 4.9
    }) {
        title
        description
        rating
    }
}
```

The above is a simple example of creating a Book using an insert mutation. Additionally, we can see that much like the Query functions, we can select the fields we want to be returned here.

The generated insert mutation returns the same type it creates, in this case, a Book type. So we can easily include all the fields as a selection set so that we can return them.

More specifically, the return type is of type `[Book]`. So we can create and return multiple books at once.

## Update

Updates are distinct from Inserts in several ways. Firstly, it relies on a query to select the correct document or documents to update. Secondly, it uses a different payload system.

Update filters use the same format and types from the Query system. Hence, it easily transferable.

The structure of the generated update mutation for a `Book` type is given below:
```graphql
mutation {
    update_Book(docID: [ID!], filter: BookFilterArg, input: BookMutationInputArg) [Book]
}
```

See the structure and syntax of the filter query above. You can also see an additional field `docID`, which will supersede the `filter`; this makes it easy to update a single document or set of documents by their IDs.

The generated schema exposes `docID` as `[ID!]`, but GraphQL input coercion also allows a single `ID` value in practice.

The input object type is the same for both `update_TYPE` and `add_TYPE` mutations.

Here's an example.
```json
{
    name: "John",
    rating: null
}
```

This update sets the `name` field to "John" and deletes the `rating` field value.

Once we create our update, and select which document(s) to update, we can query the new state of all documents affected by the mutation. This is because our update mutation returns the type it mutates.

A basic example is provided below:
```graphql
mutation {
    update_Book(docID: ["123"], input: {name: "John"}) {
        _docID
        name
    }
}

```

Here, we can see that after applying the mutation, we return the `_docID` and `name` fields. We can return any field from the document (not just the updated ones). We can even return and filter on related types.

Beyond updating by an ID or IDs, we can use a query filter to select which fields to apply our update to. This filter works the same as the queries.

```graphql
mutation {
    update_Book(filter: {rating: {_leq: 1.0}}, input: {rating: 1.5}) {
        _docID
        rating
        name
    }
}
```

Here, we select all documents with a rating less than or equal to 1.0, update the rating value to 1.5, and return all the affected documents `_docID`, `rating`, and `name` fields.

For additional filter details, see the above `Query Block` section.


## Delete

Deleting mutations allow developers and users to remove objects from collections. You can delete using specific document IDs, or a filter statement.

The document selection interface is identical to the `Update` system. Much like the update system, we can return the fields of the deleted documents.

The structure of the generated delete mutation for a `Book` type is given below:
```graphql
mutation {
    delete_Book(docID: [ID!], filter: BookFilterArg) [Book]
}
```

Here, we can delete a document with ID "123":
```graphql
mutation {
    delete_User(docID: ["123"]) {
        _docID
        name
    }
}
```

This will delete the specific document, and return the `_docID` and `name` for the deleted document.

As with updates, the generated schema exposes `docID` as `[ID!]`, while GraphQL input coercion also allows a single `ID` value in practice.

DefraDB uses a soft delete system. When a document is deleted, it is logically marked as deleted rather than physically removed from the database. Deleted documents can be retrieved using the `showDeleted` query argument.

Similar to the Update system, you can use a filter to select which documents to delete, as shown below:

```graphql
mutation {
    delete_User(filter: {rating: {_gt: 3}}) {
        _docID
        name
    }
}
```

Here, we are deleting all the matching documents (documents with a rating greater than 3).

## Upsert

Upsert mutations combine update and insert behavior into a single operation. If a document matching the provided filter is found, it will be updated with the `update` input. If no matching document is found, a new document will be created using the `add` input. The filter must match at most one document.

The structure of the generated upsert mutation for a `Book` type is given below:
```graphql
mutation {
    upsert_Book(filter: BookFilterArg!, add: BookMutationInputArg!, update: BookMutationInputArg!) [Book]
}
```

Here is an example that upserts a book by title:
```graphql
mutation {
    upsert_Book(
        filter: {title: {_eq: "Painted House"}},
        add: {title: "Painted House", rating: 4.9},
        update: {rating: 4.9}
    ) {
        _docID
        title
        rating
    }
}
```

If a book with the title "Painted House" exists, its rating will be updated to 4.9. Otherwise, a new book will be created with the provided `add` input.

It is highly recommended to add an index on the fields used in the upsert filter for best performance.
