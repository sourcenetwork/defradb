---
sidebar_label: Schema Relationship Guide
sidebar_position: 50
---
# A Guide to Schema Relationship in DefraDB

## Overview
Schema systems allow developers to enforce a structure on a given object type or database, which might be represented as rows in a SQL-based database or documents in a no SQL-based database. This enables developers to understand the structure of these objects so they can have type safety, structure safety, or the ability to enforce certain invariance or priorities syntactically or semantically.

A developer can easily enforce two separate local schema types for two objects. However, many variables need to be properly handled when it comes to the mechanism of cross-schema relationships. For example, when creating relationships between instances of documents, the developer has to design these relationships in advance with certain considerations.

Different types of relationships exist between documents and schemas. It is generally categorized based on the number of types as follows:

One-to-One Relationship: One document has a single reference to another document of a different type and vice versa.

One-to-Many/Many-to-One Relationship: One document can be referenced by many documents. For example, an author has many books and each of these books refer to one author.

Many-to-Many Relationship: Many-to-many allows developers to correlate a set of schema objects on one side to a whole set of schema objects on the other side. For example, a defined set of genres has a series of books. An intermediary relationship is created such that many correlated books have various genres. Conversely, various genres have many correlated books. Note: Many-to-many relationship is currently not supported by DefraDB, but it can be implemented through other techniques.

The developer will design and structure these relationships within the actual data of the database. Conversely, with managed relationships, the database assumes some of the responsibility of designing and maintaining the data. It depends on how the developer designs the primary and foreign keys, and how they correlate from the respective relationship model.

## How It Works  

DefraDB supports managed relationships but not unmanaged relationships, i.e., the database is responsible for accurately correlating and associating documents to their respective relationship types, primary keys, and foreign keys. The developer will be explicit about the kind of correlation they are choosing, i.e., one-to-one, one-to-many, or many-to-many; but does not have to be explicit in defining their schemas. However, the developer is not responsible for defining the field that manages the foreign keys, how that relates to the primary keys of the respective types, which side of the document is responsible for maintaining the relationship, etc. This is because the side that holds on to the foreign key is decided based on the type of relationship. In general, when querying over a relationship, the developer will define a join operation (which will allow querying from two separate tables or collections) and find a way to correlate the results into a single set of values. It should be noted that it is more efficient to query from the primary side to the secondary side.

By default, for unmanaged databases (e.g., SQL model) that has normalized tables and uses a left join, the developer will define which field on which table correlates to which field on another table. This is not the case with managed databases like DefraDB, where a type join is used in place of any other join. Type join systems reduce the complexities when defining the row join or the field join as this is automatically handled by the database.

Managing relationships for schemas is both easy and powerful in DefraDB. This is why in one-to-one relationships, DefraDB can automatically configure which side is the primary side of the relationship and can define how a developer queries different types.

However, there are some shortcomings in how Defra handles these relationships. This is because, firstly, documents in Defra are self-describing, and secondly, content identifiers of the documents are used to create the primary keys. Eventually, it becomes a little different when compared to regular databases where a primary key is an auto-incrementing integer or a randomly generated UID. Therefore, because of the operations taking place between the documents and Merkle security, the developer must keep certain causality mechanisms in mind. An example for this is a primary relationship between an "author" and "book", where "author” is the primary side in that system. The developer will either know the doc key of the book before they create the relationship, or they will create a primary key i.e., "author", which will reference "book", and then update that document once they have created the "book" to build the relationship.

When it comes to a one-to-many relationship, there is no primary side, i.e., the developer has no option of choosing which side is primary or  secondary. In this relationship type, the “many” type is the primary, and the “one” type is the secondary. Therefore, in the example of "author" to "books", if one "author" has many "books”, the book type holds the reference to the foreign key of the author type. This allows DefraDB to keep single fields on the respective type, otherwise "author" will have an array of values, thereby complicating the structure and breaking the normalization mechanism of databases.

Note: When adding related types, the developer must add both types or all related types at the same time, i.e., define all the types within the Schema Definition Language (SDL), and send them as a single schema add operation, or the database will not understand the correlated types.

With respect to filtering on related types, for both one-to-one and one-to-many relationships, the developer can filter on parent objects, which have different semantics than filtering on the child objects or the related object. Filtering on the parent object only returns the parent object if the related type matches the filter. However, filtering on the related type returns the parent regardless, but it won't return the related sub-type unless it matches the filter. For example, if we ask for authors that have books in a certain genre, it will only return authors with those sub-values for the books. If we apply a filter to the sub-type, it will return all the authors, but only return the books that match that filter.

Note: In managed relationships, the developer can also apply filtering on the values of the sub-types.

Currently, DefraDB does not support many-to-many relationships as it relates to content-identifiable data structures, self-verifying documents, and other variables, making its implementation complicated. An intermediary table, often referred to as the junction table, is used to correlate the primary key of one side of the many to the primary key of the other side of the many. Also, when creating a relationship, there is implicit data created which becomes complicated for the purposes of self-describing, self-authenticated data structures, and privacy-preserving, ownership verification of data.

Defra also does not support cascading deletes. In cascading deletes, if the developer deletes one side of a relationship, they can define a side effect or a cascade that will affect other documents, rows, or tables. While Defra does not support this feature currently, it may be included in a future version update.

## Guidelines

The following pointers provide a concrete guide on how to implement various definitions for the two managed relationship types: one-to-one and one-to-many, as well as the process of creating, updating and querying documents for the respective relationship types.

### Guidelines for One-to-One Type

1. Define the Schema and Add Types - Here, an example of two schema types - "user" and "address", where "user" has one "address" and "address" has one "user", thereby establishing a one-to-one relationship between them. The user type contains the name, age, and username, while the address type contains the street name, street number, and country. “user" is specified as the primary side of this relationship because it is more likely that the developer will query from a user to find their address rather than querying from an address to find its respective user.

    Once these schemas are loaded into the database, it will automatically create the necessary foreign keys in the respective types.

```graphql
type User {
  name: String
  username: String
  age: Int
  address: Address @primary
}

type Address {
  streetNumber: String
  streetName: String
  country: String
  user: User
}
```

2. Create and Update Mutations - Creating documents in Defra is based on the nature of how these documents need to exist because of their content identifiable structure. The developer has to first create the primary side, then the secondary side, and then update the primary side.

    In the above example, the developer will first create the "user". This is the first mutation.

```graphql
mutation {
  create_Address(input: {streetNumber: "123", streetName: "Test road", country: "Canada"}) {
  	_key
  }
}
```

```graphql
mutation {
  create_User(input: {name: "Alice", username: "awesomealice", age: 35, address_id: "bae-be6d8024-4953-5a92-84b4-f042d25230c6"}) {
  	_key
  }
}
```

Note: Currently, the developer must create the secondary side of the relation (`Address`) first followed by the primary side with the secondary id (`address_id`) included, but in a future version of Defra, this can be done in either order.

3. Querying Types - After creating the required documents, the developer has to send a query request from the primary side. Therefore, in the above example, it will ask for the three respective fields of the "user", and it will also have the embedded address type in the selection set. As the developer will query from the "user" into the "address", and as defined above, the "user" is the primary type, this lookup of "user" into "address" will be an efficient lookup that will only require a single point. A single point lookup means that it won't incur a table scan. This is explained in the query below:

```graphql
query {
    User {
        name
        username
        age
        Address {
            streetNumber
            streetName
            country
        }
    }
}

```

```graphql
query {
    Address {
        streetNumber
        streetName
        country
        User {
            name
            username
            age
        }
    }
}

```

```graphql
query {
    User (filter: {Address: {country: "Canada"}}) {
        name
        username
        age
        Address {
            streetNumber
            streetName
            country
        }
    }
}

```

Going from the secondary into the primary will be a more expensive query operation because it requires a table scan looking for the correlated user for this address.

Note: Defra supports queries from both sides, regardless of which side is the primary or secondary, i.e., the developer can query in the reverse direction.

## Guidelines for One-to-Many Type

1. Define the Schema and Add Types - For the one-to-many relationship, two types are defined, for example, "author" and "book".  The author type has a name, a birth date, and authored books. This is going to be a one-to-many relationship into the book type. The book type has a name, a description, a single genre string, and the author to which it is related.  So "author" is the one, and the "book" is the many. 


```graphql
# schema.graphql

type Author {
    name: String
    dateOfBirth: DateTime
    authoredBooks: [Book]
}

type Book {
    name: String
    description: String
    genre: String
    author: Author
}

```


```bash
defradb client schema add -f schema.graphql
```

2. Create Documents - In this step, first the "one" type from the one-to-many type needs to be created. Therefore, in the above-mentioned example, a blank author type will be created. Once "author" is created, then the related books published by the author will be created. 

    Note: Currently Defra only supports creating one type at a time, but the developer can repeat this as many times as required.

```graphql
mutation {
    create_Author(input: {name: "Saadi", dateOfBirth: "1210-07-23T03:46:56.647Z"}) {
    	_key
    }
}
```


```graphql
mutation {
  	create_Book(input: {name: "Gulistan", genre: "Poetry", author_id: "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4"}) {
      	_key
    }
}
```


```graphql
mutation {
  	update_Author(id: "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4", input: {name: "Saadi Shirazi"}) {
      	_key
    }
}
```


```graphql
mutation {
  	update_Book(filter: {name: {_eq: "Gulistan"}}, input: {description: "Persian poetry of ideas"}) {
      	_key
    }
}
```
This demonstrates that the developer should define or correlate the two types "author" and "book" using the same property of querying from the primary type to the secondary type and that in a one-to-many relationship, the "many" side is always the primary. This means that the developer has to store the related ID on the primary side, i.e., the "many" side. So, the "book" in the author-book relationship needs to hold onto the ID of the related type, i.e., the author ID. 

Note: The developer can create as many books they require by using this pattern. 

3. Querying Types - There are two directions in which the developer can run a query, i.e., secondary-to-primary or primary-to-secondary. In the first example, we are sending a query request from the "author", i.e., the secondary. It asks for all the authors in the author collection, their names, their ages, and the other fields that are on the author type, including the related book field. This book field forms the "many" end of the relationship and based on how it is defined under Point 1 above, is an array of book. Hence, it returns multiple values as an array as opposed to returning a single object, as in the case of one-to-one relationships.

    As a result, on querying a related type for a one-to-many, the developer can also sub-filter the related type. In other words, if an author has a certain number of books, we can filter the author by their name but return the books that are part of a particular genre. So, in the author-to-book direction, the developer can filter on two different levels - filter on the top level of the actual author type of the collection or filter on the book level - both having two different implications. This is further explained in the Query Specifications document. 

```graphql
query {
    Author {
      name
      dateOfBirth
      authoredBooks {
          name
          genre
          description
        }
    }
}
```


```json
// Results:
[
  {
    "name": "Saadi Shirazi",
    "dateOfBirth":  "1210-07-23T03:46:56.647Z",
    "authoredBooks": [
    	{
            "name": "Gulistan",
            "genre": "Poetry",
          	"description": "Persian poetry of ideas"
    	},
    	{
            "name":  "Bustan",
            "genre": "Poetry"
    	}
    ]
  }
]
```

```graphql
query {
    Book {
        name
        genre
        Author {
          name
          dateOfBirth
        }
    }
}
```

```json
// Results:
[
  {
    "name": "Gulistan",
    "genre": "Poetry",
    "Author": {
      "name": "Saadi Shirazi",
      "dateOfBirth":  "1210-07-23T03:46:56.647Z",
    }
  },
  {
    "name": "Bustan",
    "genre": "Poetry",
    "Author": {
      "name": "Saadi Shirazi",
      "dateOfBirth":  "1210-07-23T03:46:56.647Z",
    }
  }
]
```

```graphql
query {
    Author {
        name
        dateOfBirth
        authoredBooks(filter: {name: {_eq: "Gulistan"}}) {
            name
            genre
        }
    }
}
```

```json
// Results:
[
  {
    "name": "Saadi Shirazi",
    "dateOfBirth":  "1210-07-23T03:46:56.647Z",
    "authoredBooks": [
    	{
            "name": "Gulistan",
            "genre": "Poetry"
    	}
    ]
  }
]
```

```graphql
query {
	# Filters on the parent object can reference child fields
	# even if they are not requested.
    Author(filter: {authoredBooks: {name: {_eq: "Gulistan"}}}) {
        name
        dateOfBirth
    }
}
```

```json
// Results:
[
  {
    "name": "Saadi Shirazi",
    "dateOfBirth":  "1210-07-23T03:46:56.647Z"
  }
]
```

Note:
The book-to-author query is included to demonstrate that the developer can query from both sides of the relationship.

The various mechanisms, filtering, and rendering properties of one-to-one apply to this section as well. Also, the ability to filter on related types in one-to-many also applies to one-to-one relationships.

## Current Limitations and Future Outlook

The notable deficiencies of the current system are as follows:

* It does not support many-to-many relationships.

* It requires multiple mutations to sort all the related types.

* It does not support random joins, i.e., currently, unmanaged relationships are not supported.

The above limitations will be eliminated in future version updates of DefraDB. Our team is also working on secondary indexes, where the developer can make queries from either side, thereby improving the performance of querying from the secondary into the primary. This is almost as efficient as querying from the primary side using point lookup as opposed to a table scan