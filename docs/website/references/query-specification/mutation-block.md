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
    create_Book(input: createBookInput) [Book]
}
```

The above example displays the general structure of an insert mutation. You call the `create_TYPE` mutation, with the given input.

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
    create_Book(input: {
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
    update_Book(dockey: ID, filter: BookFilterArg, input: updateBookInput) [Book]
}
```

See the structure and syntax of the filter query above. You can also see an additional field `id`, thawhich will supersede the `filter`; this makes it easy to update a single document by a given ID.

The input object type is the same for both `update_TYPE` and `create_TYPE` mutations.

Here's an example.
```json
{
    name: "John",
    rating: nil
}
```

This update sets the `name` field to "John" and deletes the `rating` field value.

Once we create our update, and select which document(s) to update, we can query the new state of all documents affected by the mutation. This is because our update mutation returns the type it mutates.

A basic example is provided below:
```graphql
mutation {
    update_Book(dockey: '123', input: {name: "John"}) {
        _key
        name
    }
}

```

Here, we can see that after applying the mutation, we return the `_key` and `name` fields. We can return any field from the document (not just the updated ones). We can even return and filter on related types.

Beyond updating by an ID or IDs, we can use a query filter to select which fields to apply our update to. This filter works the same as the queries.

```graphql
mutation {
    update_Book(filter: {rating: {_le: 1.0}}, input: {rating: 1.5}) {
        _key
        rating
        name
    }
}
```

Here, we select all documents with a rating less than or equal to 1.0, update the rating value to 1.5, and return all the affected documents `_key`, `rating`, and `name` fields.

For additional filter details, see the above `Query Block` section.


## Delete

Deleting mutations allow developers and users to remove objects from collections. You can delete using specific Document Keys, a list of doc keys, or a filter statement.

The document selection interface is identical to the `Update` system. Much like the update system, we can return the fields of the deleted documents.

The structure of the generated delete mutation for a `Book` type is given below:
```graphql
mutation {
    delete_Book(dockey: ID, ids: [ID], filter: BookFilterArg) [Book]
}
```

Here, we can delete a document with ID '123':
```graphql
mutation {
    delete_User(dockey: '123') {
        _key
        name
    }
}
```

This will delete the specific document, and return the `_key` and `name` for the deleted document.

DefraDB currently uses a Hard Delete system, which means that when a document is deleted, it is completely removed from the database.

Similar to the Update system, you can use a filter to select which documents to delete, as shown below:

```graphql
mutation {
    delete_User(filter: {rating: {_gt: 3}}) {
        _key
        name
    }
}
```

Here, we are deleting all the matching documents (documents with a rating greater than 3).
