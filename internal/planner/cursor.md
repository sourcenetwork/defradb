# Cursor Pagination

## GraphQL shape

```graphql
query {
  _cursor {
    users: User(first: 10, after: $cursor, order: {age: ASC}) {
      name
      age
    }
    _pageInfo {
      hasNext
      hasPrev
      startCursor
      endCursor
    }
  }
}
```

`_cursor` wraps exactly one ordered collection query plus optional
`_pageInfo`. The inner collection field can be aliased independently of
the outer `_cursor` field, and the final response is rendered through a
wrapper `DocumentMapping` so normal GraphQL alias behavior is preserved.

## Execution model

The mapper parses `_cursor` into a normal collection select plus cursor
arguments. Planner analysis still builds the ordinary child plan for the
collection, including filtering, joins, scanning, and ordering. The
cursor layer sits above that ordered child plan and is only responsible
for:

- interpreting `first`/`after` and `last`/`before`
- finding the current page boundaries
- computing `_pageInfo`
- encoding the returned page boundaries back into cursors

`cursorNode` does not sort or filter rows itself. It assumes planner
expansion has already produced a child plan whose iteration order matches
the requested cursor order and whose index usage has been validated.

## Cursor tokens

A cursor token always identifies a boundary row by DocID. When the
requested order can be resumed through an index seek, the token also
carries the ordered key values from that row. That extra key data lets
the system resume from a stable boundary without using offsets.

Older DocID-only tokens are still accepted, but they do not always carry
enough information to reconstruct a safe index seek position.

## Forward and backward paths

For forward pagination, `cursorNode` returns rows after `after` up to
`first`. If planner enabled index seeking, the child scan starts from
the seek position. Otherwise the node skips rows in logical order until
the boundary row has been passed.

For backward pagination, the planner first decides whether the `before`
token contains enough key data to seek into the index safely:

- If it does, the child scan is rewired into reverse iteration,
  `cursorNode` buffers the requested page, and then restores the
  requested logical order before returning rows.
- If it does not, the node falls back to draining rows in logical order
  up to `before` and then taking the tail of that stream.

That fallback is important for legacy DocID-only cursors and for any
cursor that does not encode a usable index boundary tuple.

## Page info and returned cursors

Both directions probe one extra row beyond the requested page size to
compute `hasNext` and `hasPrev`. While iterating, `cursorNode` tracks the
first and last visible rows of the current page so
`_pageInfo.startCursor` and `_pageInfo.endCursor` are encoded from the
actual rows returned to the user, not from the input cursor.
