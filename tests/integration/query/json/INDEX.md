# Index: `tests/integration/query/json`

## Overview

This folder contains integration tests for querying JSON-typed fields in DefraDB. It covers the full range of filter operators available on JSON fields: equality (`_eq`, `_neq`), set membership (`_in`, `_nin`), numeric comparison (`_gt`, `_geq`, `_lt`, `_leq`), string pattern matching (`_like`, `_nlike`), array quantifiers (`_all`, `_any`, `_none`), and aggregate queries (COUNT). Tests exercise top-level JSON values of every JSON type (number, string, boolean, object, array, null), nested object paths, and the type-safety constraints that reject invalid filter operand types.

## Test Index

### `with_aggregate_test.go`

COUNT aggregate filtered by a JSON field condition.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryJSON_WithAggregateFilter_Succeeds` | 21-67 | COUNT aggregate with a JSON field filter correctly counts matching documents. |

### `with_all_test.go`

`_all` quantifier filter on JSON array fields, covering mixed-type arrays and nested array behavior.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryJSON_WithAllFilterWithAllTypes_ShouldFilter` | 21-91 | _all filter on a JSON array field returns only documents where all elements are non-null. |
| `TestQueryJSON_WithAllFilterAndNestedArray_ShouldFilter` | 93-149 | _all filter on a JSON array checks only top-level elements, not nested array contents. |

### `with_any_test.go`

`_any` quantifier filter on JSON array fields, covering mixed-type arrays and nested array behavior.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryJSON_WithAnyFilterWithAllTypes_ShouldFilter` | 21-85 | _any filter on a JSON array field returns documents that contain at least one null element. |
| `TestQueryJSON_WithAnyFilterAndNestedArray_ShouldFilter` | 87-155 | _any filter on a JSON array matches only top-level scalar elements, not values inside nested arrays. |

### `with_eq_test.go`

`_eq` filter on JSON fields: object equality, compound `_and` conditions, deeply nested objects, null values, mixed JSON types, and nested path equality.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryJSON_WithEqualFilterWithObject_ShouldFilter` | 21-73 | _eq filter on a JSON field matches a document whose value equals the given object. |
| `TestQueryJSON_WithCompoundFilterCondition_ShouldFilter` | 75-133 | Compound _and filter on two nested JSON field conditions returns only the matching document. |
| `TestQueryJSON_WithEqualFilterWithNestedObjects_ShouldFilter` | 135-187 | _eq filter on a JSON field matches a document with deeply nested objects and arrays. |
| `TestQueryJSON_WithEqualFilterWithNullValue_ShouldFilter` | 189-229 | _eq null filter on a JSON field returns only documents whose field value is null. |
| `TestQueryJSON_WithEqualFilterWithAllTypes_ShouldFilter` | 231-291 | _eq object filter on a JSON field matches only the document stored as an object, skipping other JSON types. |
| `TestQueryJSON_WithEqualFilterWithObjectValueOnNestedPath_ShouldFilter` | 293-346 | _eq filter on a nested JSON path matches documents where the nested value equals the given object. |

### `with_geq_test.go`

`_geq` filter on JSON numeric fields: equal and greater values, null semantics, nested paths, and invalid operand types that return errors.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryJSON_WithGreaterEqualFilterWithEqualValue_ShouldFilter` | 21-63 | _geq filter on a JSON numeric field returns documents whose value is greater than or equal to the threshold. |
| `TestQueryJSON_WithGreaterEqualFilterWithGreaterValue_ShouldFilter` | 65-107 | _geq filter on a JSON numeric field excludes documents whose value is strictly less than the threshold. |
| `TestQueryJSON_WithGreaterEqualFilterWithNullValue_ShouldFilter` | 109-154 | _geq null filter on a JSON field returns all documents because every value is >= null. |
| `TestQueryJSON_WithGreaterEqualFilterWithNestedEqualValue_ShouldFilter` | 156-198 | _geq filter on a nested JSON numeric field returns documents whose nested value meets the threshold. |
| `TestQueryJSON_WithGreaterEqualFilterWithNestedGreaterValue_ShouldFilter` | 200-242 | _geq filter on a nested JSON path excludes documents whose nested value is below the threshold. |
| `TestQueryJSON_WithGreaterEqualFilterWithNestedNullValue_ShouldFilter` | 244-298 | _geq null filter on a nested JSON path returns all documents including those with null or missing nested values. |
| `TestQueryJSON_WithGreaterEqualFilterWithBoolValue_ReturnsError` | 300-336 | _geq filter with a boolean value on a JSON field returns an unexpected-type error. |
| `TestQueryJSON_WithGreaterEqualFilterWithStringValue_ReturnsError` | 338-374 | _geq filter with a string value on a JSON field returns an unexpected-type error. |
| `TestQueryJSON_WithGreaterEqualFilterWithObjectValue_ReturnsError` | 376-412 | _geq filter with an object value on a JSON field returns an unexpected-type error. |
| `TestQueryJSON_WithGreaterEqualFilterWithArrayValue_ReturnsError` | 414-450 | _geq filter with an array value on a JSON field returns an unexpected-type error. |
| `TestQueryJSON_WithGreaterEqualFilterWithAllTypes_ShouldFilter` | 452-512 | _geq filter on a JSON field matches only the numeric document when mixed JSON types are stored. |

### `with_gt_test.go`

`_gt` filter on JSON numeric fields: greater and lesser values, null semantics, nested paths, and invalid operand types that return errors.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryJSON_WithGreaterThanFilterBlockWithGreaterValue_ShouldFilter` | 21-65 | _gt filter on a JSON numeric field returns the document whose value exceeds the threshold. |
| `TestQueryJSON_WithGreaterThanFilterBlockWithLesserValue_ShouldFilter` | 67-106 | _gt filter on a JSON numeric field returns no documents when no value exceeds the threshold. |
| `TestQueryJSON_WithGreaterThanFilterBlockWithNullFilterValue_ShouldFilter` | 108-149 | _gt null filter on a JSON field returns documents with any non-null value. |
| `TestQueryJSON_WithGreaterThanFilterBlockWithNestedGreaterValue_ShouldFilter` | 151-197 | _gt filter on a nested JSON numeric field returns the document whose nested value exceeds the threshold. |
| `TestQueryJSON_WithGreaterThanFilterBlockWithNestedLesserValue_ShouldFilter` | 199-238 | _gt filter on a nested JSON numeric field returns no documents when no nested value exceeds the threshold. |
| `TestQueryJSON_WithGreaterThanFilterBlockWithNestedNullFilterValue_ShouldFilter` | 240-281 | _gt null filter on a nested JSON path returns documents with a non-null nested value. |
| `TestQueryJSON_WithGreaterThanFilterBlockWithBoolValue_ReturnsError` | 283-320 | _gt filter with a boolean value on a JSON field returns an unexpected-type error. |
| `TestQueryJSON_WithGreaterThanFilterBlockWithStringValue_ReturnsError` | 322-359 | _gt filter with a string value on a JSON field returns an unexpected-type error. |
| `TestQueryJSON_WithGreaterThanFilterBlockWithObjectValue_ReturnsError` | 361-398 | _gt filter with an object value on a JSON field returns an unexpected-type error. |
| `TestQueryJSON_WithGreaterThanFilterBlockWithArrayValue_ReturnsError` | 400-437 | _gt filter with an array value on a JSON field returns an unexpected-type error. |
| `TestQueryJSON_WithGreaterThanFilterWithAllTypes_ShouldFilter` | 439-499 | _gt filter on a JSON field matches only the numeric document when mixed JSON types are stored. |

### `with_in_test.go`

`_in` filter on JSON fields matching a list of object values.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryJSON_WithInFilter_ShouldFilter` | 21-67 | _in filter on a JSON field returns only the document whose value is in the given list. |

### `with_leq_test.go`

`_leq` filter on JSON numeric fields: equal and lesser values, null semantics, nested paths, and invalid operand types that return errors.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryJSON_WithLesserEqualFilterWithEqualValue_ShouldFilter` | 21-63 | _leq filter on a JSON numeric field returns documents whose value is less than or equal to the threshold. |
| `TestQueryJSON_WithLesserEqualFilterWithLesserValue_ShouldFilter` | 65-107 | _leq filter on a JSON numeric field excludes documents whose value exceeds the threshold. |
| `TestQueryJSON_WithLesserEqualFilterWithNullValue_ShouldFilter` | 109-150 | _leq null filter on a JSON field returns only documents with a null or absent field value. |
| `TestQueryJSON_WithLesserEqualFilterWithNestedEqualValue_ShouldFilter` | 152-194 | _leq filter on a nested JSON numeric field returns documents whose nested value is at or below the threshold. |
| `TestQueryJSON_WithLesserEqualFilterWithNestedLesserValue_ShouldFilter` | 196-238 | _leq filter on a nested JSON path excludes documents whose nested value exceeds the threshold. |
| `TestQueryJSON_WithLesserEqualFilterWithNestedNullValue_ShouldFilter` | 240-290 | _leq null filter on a nested JSON path returns documents with a null or absent nested value. |
| `TestQueryJSON_WithLesserEqualFilterWithBoolValue_ReturnsError` | 292-328 | _leq filter with a boolean value on a JSON field returns an unexpected-type error. |
| `TestQueryJSON_WithLesserEqualFilterWithStringValue_ReturnsError` | 330-366 | _leq filter with a string value on a JSON field returns an unexpected-type error. |
| `TestQueryJSON_WithLesserEqualFilterWithObjectValue_ReturnsError` | 368-404 | _leq filter with an object value on a JSON field returns an unexpected-type error. |
| `TestQueryJSON_WithLesserEqualFilterWithArrayValue_ReturnsError` | 406-442 | _leq filter with an array value on a JSON field returns an unexpected-type error. |
| `TestQueryJSON_WithLesserEqualFilterWithAllTypes_ShouldFilter` | 444-504 | _leq filter on a JSON field matches only the numeric document when mixed JSON types are stored. |

### `with_like_test.go`

`_like` pattern filter on JSON string fields, ignoring non-string JSON values.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryJSON_WithLikeFilter_ShouldFilter` | 21-80 | _like filter on a JSON string field matches only documents whose string value fits the pattern. |

### `with_lt_test.go`

`_lt` filter on JSON numeric fields: greater and lesser values, null semantics, nested paths, and invalid operand types that return errors.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryJSON_WithLesserThanFilterBlockWithGreaterValue_ShouldFilter` | 21-65 | _lt filter on a JSON numeric field returns the document whose value is below the threshold. |
| `TestQueryJSON_WithLesserThanFilterBlockWithLesserValue_ShouldFilter` | 67-106 | _lt filter on a JSON numeric field returns no documents when no value is below the threshold. |
| `TestQueryJSON_WithLesserThanFilterBlockWithNullFilterValue_ShouldFilter` | 108-145 | _lt null filter on a JSON field returns no documents because no value is less than null. |
| `TestQueryJSON_WithLesserThanFilterBlockWithNestedGreaterValue_ShouldFilter` | 147-193 | _lt filter on a nested JSON numeric field returns the document whose nested value is below the threshold. |
| `TestQueryJSON_WithLesserThanFilterBlockWithNestedLesserValue_ShouldFilter` | 195-234 | _lt filter on a nested JSON path returns no documents when no nested value is below the threshold. |
| `TestQueryJSON_WithLesserThanFilterBlockWithNestedNullFilterValue_ShouldFilter` | 236-273 | _lt null filter on a nested JSON path returns no documents because no value is less than null. |
| `TestQueryJSON_WithLesserThanFilterBlockWithBoolValue_ReturnsError` | 275-312 | _lt filter with a boolean value on a JSON field returns an unexpected-type error. |
| `TestQueryJSON_WithLesserThanFilterBlockWithStringValue_ReturnsError` | 314-351 | _lt filter with a string value on a JSON field returns an unexpected-type error. |
| `TestQueryJSON_WithLesserThanFilterBlockWithObjectValue_ReturnsError` | 353-390 | _lt filter with an object value on a JSON field returns an unexpected-type error. |
| `TestQueryJSON_WithLesserThanFilterBlockWithArrayValue_ReturnsError` | 392-429 | _lt filter with an array value on a JSON field returns an unexpected-type error. |
| `TestQueryJSON_WithLesserThanFilterWithAllTypes_ShouldFilter` | 431-491 | _lt filter on a JSON field matches only the numeric document when mixed JSON types are stored. |

### `with_neq_test.go`

`_neq` filter on JSON fields: object inequality, nested objects, null values, and nested path comparisons against scalar types.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryJSON_WithNotEqualFilterWithObject_ShouldFilter` | 21-75 | _neq filter on a JSON field excludes the document whose value equals the given object. |
| `TestQueryJSON_WithNotEqualFilterWithNestedObjects_ShouldFilter` | 77-129 | _neq filter on a JSON field excludes the document with deeply nested objects matching the given value. |
| `TestQueryJSON_WithNotEqualFilterWithNullValue_ShouldFilter` | 131-171 | _neq null filter on a JSON field returns only documents whose field value is non-null. |
| `TestQueryJSON_WithNeFilterAgainstNumberField_ShouldFilter` | 173-225 | _neq filter on a nested JSON numeric field returns documents whose nested value differs from the given number. |
| `TestQueryJSON_WithNeFilterAgainstStringField_ShouldFilter` | 227-279 | _neq filter on a nested JSON string field returns documents whose nested value differs from the given string. |
| `TestQueryJSON_WithNeFilterAgainstBooleanField_ShouldFilter` | 281-333 | _neq filter on a nested JSON boolean field returns documents whose nested value differs from the given boolean. |
| `TestQueryJSON_WithNeFilterAgainstNullField_ShouldFilter` | 335-387 | _neq null filter on a nested JSON path returns only documents with a non-null nested value. |
| `TestQueryJSON_WithNotEqualFilterWithNestedObject_ShouldFilter` | 389-443 | _neq filter on a nested JSON path excludes the document whose nested value equals the given object. |

### `with_nin_test.go`

`_nin` filter on JSON fields excluding documents whose value matches any entry in the given list.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryJSON_WithNotInFilter_ShouldFilter` | 21-67 | _nin filter on a JSON field returns only documents whose value is not in the given list. |

### `with_nlike_test.go`

`_nlike` pattern filter on JSON string fields, returning all non-matching documents including non-string JSON values.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryJSON_WithNotLikeFilter_ShouldFilter` | 21-93 | _nlike filter on a JSON field returns all documents whose string value does not match the pattern. |

### `with_none_test.go`

`_none` quantifier filter on JSON array fields, verifying that only top-level elements are checked.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryJSON_WithNoneFilter_ShouldFilter` | 21-61 | _none filter on a JSON array field returns only documents where no element equals null. |
| `TestQueryJSON_WithNoneFilterAndNestedArray_ShouldFilter` | 63-119 | _none filter on a JSON array checks only top-level elements and excludes documents where any element equals the value. |
