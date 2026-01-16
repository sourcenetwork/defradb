---
sidebar_label: Filtering
sidebar_position: 50
---
# Filtering

Filtering is used to screen data entries containing the specified fields and predicates (including compound predicates) out of a collection of documents using conditional keywords like `_and`, `_or`, `_not`. To accomplish this, the `filter` keyword can be applied as an argument to root level fields and subfields.

An empty `filter` object is equivalent to no filters being applied. Hence, the output will return all books. The following example displays an empty filter being applied on the root level field.

```graphql
{
	Books(filter: {}) {
		title
		genre
		description
	}
}
```

Some filtering options depend on the available indexes on a field. However, we will not be discuss them in this section.

To apply a filter to a specific field, we can specify it within the filter object. The example below only returns books with the title “A Painted House”.

```graphql
{
	Books(filter: { title: { _eq: "A Painted House" }}) {
		title
		genre
		description
	}
}
```

We can apply filters to all or multiple fields available.

**NOTE:** Each additional field listed in the filter object implies to a conditional AND relation.

```graphql
{
	Books(filter: { title: {_eq: "A Painted House"}, genre: {_eq: "Thriller" }}) {
		title
		genre
		description
	}
}
```

The above query only returns books with the title “A Painted House” AND genre “Thriller”.

Filters can also be applied on subfields that have relational objects within them. For example: an object Book, with an Author field, has a many-to-one relationship to the Author object. Then we can query and filter based on the value of the Author field.

```graphql
{
	Books(filter: { genre: {_eq: "Thriller"}, author: {name: {_eq: "John Grisham"}}}) {
		title
		genre
		description
		Author {
			name
			bio
		}
	}
}
```

This query returns all books authored by “John Grisham” with the genre “Thriller”.

Filtering from the root object level, compared to the sub-object level results in different semantics. Root filters that apply to sub-objects (aka `author` section of the above query), only returns the root object type if both the root object and sub-object conditions are fulfilled. For example, if the author filter condition is satisfied, the above code snippet only returns books.

This applies to both single sub-objects and array sub-objects, i.e., if we apply a filter on a sub-object array, the output **only** returns the root object, if at least one sub-object matches the given filter instead of requiring **every** sub-object to match the query. For example, the following query will only return authors, if they have **at least** one thriller genre based book.

```graphql
{
    Authors(filter: {book: {genre: {_eq: "Thriller"}}}) {
        name
        bio
    }
}
```

Additionally, in the selection set, if we include the sub-object array we are filtering on, the filter is then implicitly applied unless otherwise specified.

In the query snippet above, let's add `books` to the selection set using the query below .
```graphql
{
	Authors(filter: {book: {genre: {_eq: "Thriller"}}}) {
        name
        bio
        books {
            title
            genre
        }
    }
}
```

Here, the `books` section will only contain books that match the root object filter, namely, `{genre: {_eq: "Thriller"}}`. If we wish to return the same authors from the above query and include *all* their books, we can add an explicit filter directly to the sub-object instead of the sub-filters.

```graphql
{
    Authors(filter: {book: {genre: {_eq: "Thriller"}}}) {
        name
        bio
        books(filter: {}) {
            title
            genre
        }
    }
}
```

In the code snippet above, the output returns authors who have at least one book with the genre "Thriller". The output also returns **all** the books written by these selected authors (not just the thrillers).

Filters applied solely to sub-objects, which are only applicable for array types, are computed independently from the root object filters.

```graphql
{
	Authors(filter: {name: {_eq: "John Grisham"}}) {
		name
		bio
		books(filter: { genre: {_eq: "Thriller" }}) {
			title
			genre
			description
		}
	}
}
```

The above query returns all authors with the name “John Grisham”, then filters and returns all the returned authors' books with the genre “Thriller”. This is similar to the previous query, but an important distinction is that it will return all the matching author objects regardless of the book's sub-object filter. 

The first query, will only return an output if there are any Thriller books written by the author “John Grisham” (using AND condition i.e., both conditions have to be fulfilled). The second query always returns all authors named “John Grisham”, and their Thriller genre books.

So far, we have only seen examples of EXACT string matches, but we can also filter using scalar value type or object fields. For e.g., booleans, integers, floating points, etc. Also, comparison operators like: Greater Than, Less Than, Equal To or Greater than, Less Than or Equal To, EQUAL can be used. 

Let's query for all books with a rating greater than or equal to 4.

```graphql
{
	Books(filter: { rating: { _gte: 4 } }) {
		title
		genre
		description
	}
}
```

**NOTE:** In the above example, the expression contains a new scalar type object `{ _gte: 4 }`.  While previously, where we had a simple string value. If a scalar type field has a filter with an object value, then that object's first and only key must be a comparison operator like `_gte`. If the filter is given a simple scalar value like “John Grisham”, “Thriller”, or FALSE, then the default operator that should be used is `_eq` (EQUAL). The following table displays a list of available operators:


| Operator | Description |
| -------- | --------    |
| `_eq`    | Equal to        |
| `_neq`   | Not Equal to        |
| `_gt`    | Greater Than        |
| `_gte`   | Greater Than or Equal to        |
| `_lt`    | Less Than        |
| `_lte`   | Less Than or Equal to        |
| `_in`    | In the List        |
| `_nin`   | Not in the List        |
| `_like`  | Like Sub-String         |
|`_nlike`  | Unlike Sub-String       |
###### Table 1. Supported operators.

The table below displays the operators that can be used for every value type:


| Scalar Type | Operators |
| -------- | -------- |
| String     | `_eq, _neq, _like, _in, _nin`     |
| Integer     | `_eq, _neq, _gt, _gte, _lt, _lte, _in, _nin`     |
| Floating Point     | `_eq, _neq, _gt, _gte, _lt, _lte, _in, _nin`     |
| Boolean     | `_eq, _neq, _in, _nin`     |
| DateTime     | `_eq, _neq, _gt, _gte, _lt, _lte, _in, _nin`     |
###### Table 2. Operators supported by Scalar types.

There are 3 types of conditional keywords, i.e, `_and`, `_or`, and `_not`. Conditional keywords like `_and` and `_or` are used when we need to apply filters on multiple fields simultaneously.  The `_not` conditional keyword only accepts an object.

The code snippet below queries all books that are a part of the Thriller genre, or have a rating between 4 to 5.

```graphql
{
    Books(
        filter: { 
            _or: [ 
                {genre: {_eq: "Thriller"}}, 
                { _and: [
                    {rating: { _gte: 4 }},
                    {rating: { _lte: 5 }},
                ]},
            ]
        }
    )
	title
	genre
	description
}
```

An important thing to note about the above query is the `_and` conditional. Even though AND is assumed, if we have two filters on the same field, we MUST specify the `_and` operator. This is because JSON objects cannot contain duplicate fields.

>**Invalid**:
`filter: { rating: { _gte: 4 }, rating { _lte: 5 } }`
>**Valid**:
`filter: { _and: [ {rating: {_gte: 4}}, {rating: {_lte: 5}} ]}`

The `_not` conditional accepts an object instead of an array.

> Filter all objects that *do not* have the genre "Thriller"
> `filter: { _not: { genre: { _eq: "Thriller" } } }`

*The`_not` operator should only be used when the available filter operators like `_neq` do not fit the use case.*
