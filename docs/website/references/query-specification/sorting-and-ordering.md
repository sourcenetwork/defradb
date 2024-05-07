---
sidebar_label: Sorting and Ordering
sidebar_position: 60
---
# Sorting and Ordering

Sorting is an integral part of any Database and Query Language. The sorting syntax is similar to filter syntax, in that we use objects, and sub-objects to indicate sorting behavior, instead of filter behavior.

The query to find all books ordered by their latest published date:
```graphql
{
    Books(order: { published_at: DESC}) {
        title
        description
        published_at
    }
}
```
The syntax indicates:
- The field we wanted to sort on `published_at`
- The direction we wanted to order by `descending`.

Sorting can be applied to multiple fields in the same query. The sort order is same as the field order in the sorted object.

The query below finds all books ordered by earliest published date and then by descending order of titles.
```graphql
{
    Books(order: { published_at: ASC, title: DESC }) {
        title
        genre
        description
    }
}
```

Additionally, you can sort sub-object fields along with root object fields.

The query below finds all books ordered by earliest published date and then by the latest authors' birthday.
```graphql
{
    Books(order: { published_at: ASC, Author: { birthday: DESC }}) {
        title
        description
        published_at
        Author {
            name
            birthday
        }
    }
}
```

Sorting multiple fields simultaneously is primarily driven by the first indicated sort field (primary field). In the query above, it is the “published_at” date. The following sort field (aka, secondary field), is used in the case that more than one record has the same value for the primary sort field. 

Assuming there are more than two sort fields, in that case, the same behavior applies, except the primary, secondary pair shifts by one element. Hence, the 2nd field is the primary, and the 3rd is the secondary, until we reach the end of the sort fields.

In case of a single sort field and objects with same value, the documents identifier (DocKey) is used as the secondary sort field by default. This is applicable regardless of the number of sort fields. As long as the DocKey is not already included in sort fields, it acts as the final tie-breaking secondary field.

If the DocKey is included in the sort fields, any field included afterwards will never be evaluated. This is because all DocKeys are unique. If the sort fields are `published_at`, `id`, and `birthday`, the `birthday` sort field will never be evaluated and should be removed from the list.

> Sorting on sub-objects from the root object is only allowed if the sub-object is not an array. If it is an array, the sort must be applied to the object field directly instead of through the root object.

*So, instead of:*
```graphql
{
    Authors(order: { name: DESC, Books: { title: ASC }}) {
        name
        Books {
            title
        }
    }
}
```
*We need:*
```graphql
{
    Authors(order: { name: DESC }) {
        name
        Books(order: { title: ASC }) {
            title
        }
    }
}
```

>Root level filters and order only apply to root object. If you allow the initial version of the query, it would be confusing if the ordering applied to the order of the root object compared to its sibling objects or if the ordering applied solely to the sub-object. 

>If you allow it, it enforces the semantics of root level sorting on array sub-objects to act as a sorting mechanism for the root object. As a result, there is no obvious way to determine which value in the array is used for the root order.

If you have the following objects in the database:
```json
 [
     "Author" {
         "name": "John Grisham",
         "books": [
            { "title": "A Painted House" },
            { "title": "The Guardians" }
         ]
     },
     "Author" {
         "name": "John Grisham",
         "books": [
            { "title": "Camino Winds" },
         ]
     },
     "Author" {
         "name": "John LeCare",
         "books": [
             { "title": "Tinker, Tailor, Soldier, Spy"}
         ]
     }
 ]
```
> and the following query
```graphql
{
    Authors(order: { name: DESC, books: { title: ASC }}) {
        name
        books {
            title
        }
    }
}
```

```graphql
Books(filter: {_id: [1]}) {
    title 
    genre
    description
}
```

> Given there are two authors with the same name (John Grisham), the sort object `(sort: { name: "desc", Books: { title: "asc" }}` would suggest we sort duplicate authors using `Books: { title: "asc" }` as the secondary sort field. However, because the books field is an array of objects, there is no single value for the title to compare easily.
>
> Therefore, sorting on array sub objects from the root field is ***strictly not allowed***.
