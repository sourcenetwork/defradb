---
sidebar_label: Aggregate Functions
sidebar_position: 115
---

# Aggregate Functions

The most common use case of grouping queries is to compute some aggregate function over the sub-group. Like the special `GROUP` field, aggregate functions are defined and returned using special fields. When used within a `groupBy` query, these fields take a `GROUP` argument specifying the target field. For example, to access the average value of all sub-group ratings, include `AVG(GROUP: {field: rating})` in your return object. The available aggregate functions are: `COUNT`, `SUM`, `AVG`, `MAX`, and `MIN`.

The aggregate function syntax uses the function name with a named argument specifying the field. Within a `groupBy` query: `AVG(GROUP: {field: $field})`, where `$field` is the field name to which the function will be applied. For top-level (non-grouped) aggregates, the syntax is `AVG($Collection: {field: $field})`.

Let us augment the previous grouped books by genre example and include an aggregate function on the sub-groups ratings.
```graphql
{
    Books(filter: {author: {name: {_like: "John%"}}}, groupBy: [genre]) {
        genre
        AVG(GROUP: {field: rating})
        GROUP {
            title
            rating
        }
    }
}
```

Here we return the average of all the ratings of the books whose authors name begins with "John" grouped by the genres.

We can also use simpler queries, without any `groupBy` clause, and still use aggregate functions. The difference is, instead of applying the aggregate function to only the sub-group, it applies it to the entire result set.

Let's simply count all the objects returned by a given filter.
```graphql
{
    COUNT(Books: {})
}
```
This returns the total number of Book objects. When using `COUNT` at the top level, it applies to the entire collection.

> Note, the special aggregate field `COUNT` does not require a target field, so instead of applying the `count` function to a field, it applies to the entire object. This is only possible with the `count` function; all the other aggregate functions must specify their target field using the `field` argument.

We can further simplify the above count query by including only the `COUNT` field. If we ***only*** return the `COUNT` field, then a single object is returned, instead of an array of objects.

DefraDB also supports applying aggregate functions to relations just like we do fields. However, only the `count` function is available directly on the related object type.


