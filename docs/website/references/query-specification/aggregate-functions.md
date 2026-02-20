---
sidebar_label: Aggregate Functions
sidebar_position: 115
---

# Aggregate Functions

The most common use case of grouping queries is to compute some aggregate function over the sub-group. Like the special `GROUP` field, aggregate functions are defined and returned using special fields. These fields prefix the target field name with the name of the aggregate function you wish to apply. If we had the field `rating`, we could access the average value of all sub-group ratings by including the special field `AVG { rating }` in our return object. The available aggregate functions and their associated scalars can be found above in `Table 3`.

The special aggregate function fields' format is the function name and the field name as its sub-elements. Specifically: `_$function { $field }`, where `$function` is the list of functions from `Table 3`, and `$field` is the field name to which the function will be applied to. E.g., applying the `max` function to the `rating` field becomes `MAX { rating }`.

Let us augment the previous grouped books by genre example and include an aggregate function on the sub-groups ratings.
```graphql
{
    Books(filter: {author: {name: {_like: "John%"}}}, groupBy: [genre]) {
        genre
        AVG {
            rating
            points
        }
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
    COUNT(Books: {filter: {rating: {_gt: 3.5}}})
}
```
This returns an array of objects that includes the respective books title, along with the repeated `COUNT` field, which is the total number of objects that match the filter.

> Note, the special aggregate field `COUNT` has no subfields selected, so instead of applying the `count` function to a field, it applies to the entire object. This is only possible with the `count` function; all the other aggregate functions must specify their target field using the correct field name selection.

We can further simplify the above count query by including only the `COUNT` field. If we ***only*** return the `COUNT` field, then a single object is returned, instead of an array of objects.

DefraDB also supports applying aggregate functions to relations just like we do fields. However, only the `count` function is available directly on the related object type.


