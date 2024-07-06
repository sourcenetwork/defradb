---
sidebar_label: Limiting and Pagination
sidebar_position: 70
---
# Limiting and Pagination

After filtering and sorting a query, we can then limit and skip elements from the returned set of objects.

Let us get the top 10 rated books:
```graphql
{
    Books(sort: {rating: DESC}, limit: 10) {
        title
        genre
        description
    }
}
```

The `limit` function accepts the maximum number of items to return from the resulting set. Next, we can `skip` elements in the set, to get the following N objects from the return set. Both these functions can be used to create a pagination system, where we have a limit on number of items per page, and can skip through pages as well.

Let's get the *next* top 10 rated books after the previous query:
```graphql
{
    Books(sort: {rating: DESC}, limit:10, offset: 10) {
        title
        genre
        description
    }
}
```

Limits and offsets can be combined to create several different pagination methods.