---
sidebar_label: Grouping
sidebar_position: 110
---
# Grouping

Grouping allows a collection of results from a query to be "grouped" into sections based on some field. These sections are called sub-groups, and are based on the equality of fields within objects, resulting in clusters of groups. Any object field may be used to group objects together. Additionally, multiple fields may be used in the group by clause to further segment the groups over multiple dimensions.

Once one or more group by fields have been selected using the `groupBy` argument, which accepts an array of length one or more, you may only access certain fields in the return object. Only the indicated `groupBy` fields and aggregate function results may be included in the result object. If you wish to access the sub-groups of individual objects, a special return field called `_group` is available. This field matches the root query type, and can access any field in the object type.

In the example below, we are querying for all the books whose author's name begins with 'John'. The results will then be grouped by genre, and will return the genre name and the sub-groups `title` and `rating`.
```graphql
{
    Books(filter: {author: {name: {_like: "John%"}}}, groupBy: [genre]) {
        genre
        _group {
            title
            rating
        }
    }
}
```

In the above example, we can see how the `groupBy` argument is provided and that it accepts an array of field names. We can also see how the special `_group` field can be used to access the sub-group elements.

It's important to note that in the above example, the only available field from the root `Book` type is the `groupBy` field `genre`, along with the special group and aggregate proxy fields.

#### Grouping on Multiple Fields
As mentioned, we can include any number of fields in the `groupBy` argument to segment the data further. Which can then also be accessed in the return object, as demonstrated in the example below:
```graphql
{
    Books(filter: {author: {name: {_like: "John%"}}}, groupBy: [genre, rating]) {
        genre
        rating
        _group {
            title
            description
        }
    }
}
```

#### Grouping on Related Objects
Objects often have related objects within their type definition indicated by the `@relation` directive on the respective object. We can use the grouping system to split results over the related object and the root type fields.

Like any other group query, we are limited in which fields we can access indicated by the `groupBy` argument's fields. If we include a subtype that has a `@relation` directive in the `groupBy` list, we can access the entire relations fields.

Only "One-to-One" and "One-to-Many" relations can be used in a `groupBy` argument.

Given a type definition defined as:
```graphql
type Book {
    title: String
    genre: String
    rating: Float
    author: Author @relation
}

type Author {
    name: String
    written: [Book] @relation
}
```

We can create a group query over books and their authors, as demonstrated in the example below:
```graphql
{
    Books(groupBy: [author]) {
        Author {
            name
        }
        _group {
            title
            genre
            rating
        }
    }
}
```

As you can see, we can access the entire `Author` object in the main return object without having to use any special proxy fields.

Group operations can include any combination, single or multiple, individual field or related object, that a developer needs.
