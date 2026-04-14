# Index: `tests/integration/query/inline_array`

## Overview

This folder contains integration tests for querying inline array fields in DefraDB. It covers all supported inline array types — boolean, integer, float, and string — in both non-null (`[Type!]`) and nillable (`[Type]`) variants. Tests verify basic query behavior (null, empty, and populated arrays), aggregate functions (AVG, COUNT, MAX, MIN, SUM) with filtering, ordering, limit, and offset arguments, element-level quantifier filters (`_all`, `_any`, `_none`), full-array equality filters (`_eq`, `_neq`), and grouping by inline array fields.

## Test Index

### `simple_test.go`

Basic querying of every inline array type in null, empty, and populated states, including nillable variants that preserve null entries.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineArrayWithBooleans_Null` | 23-53 | Querying an inline boolean array field that is null returns nil. |
| `TestQueryInlineArrayWithBooleans_EmptyList` | 54-84 | Querying an inline boolean array field that is an empty list returns an empty slice. |
| `TestQueryInlineArrayWithBooleans_NotEmpty` | 85-115 | Querying an inline boolean array field with values returns the correct boolean slice. |
| `TestQueryInlineArrayWithNillableBooleans` | 117-152 | Querying a nillable inline boolean array field returns values including null entries. |
| `TestQueryInlineArrayWithIntegers_Missing` | 154-183 | Querying an inline integer array field not present in the document returns nil. |
| `TestQueryInlineArrayWithIntegers_Null` | 185-215 | Querying an inline integer array field set to null returns nil. |
| `TestQueryInlineArrayWithIntegers_EmptyList` | 217-247 | Querying an inline integer array field set to an empty list returns an empty slice. |
| `TestQueryInlineArrayWithIntegers_NotEmptyList` | 249-279 | Querying an inline integer array field with positive values returns the correct integer slice. |
| `TestQueryInlineArrayWithNegativeIntegers_NotEmptyList` | 281-311 | Querying an inline integer array field with all negative values returns the correct slice. |
| `TestQueryInlineArrayWithMixIntegers_NotEmptyList` | 313-343 | Querying an inline integer array field with mixed positive, negative, and zero values returns correctly. |
| `TestQueryInlineArrayWithNillableInts` | 344-380 | Querying a nillable inline integer array field returns values preserving null entries. |
| `TestQueryInlineArrayWithFloats_Null` | 382-412 | Querying an inline float array field set to null returns nil. |
| `TestQueryInlineArrayWithFloats_EmptyList` | 414-444 | Querying an inline float array field set to an empty list returns an empty slice. |
| `TestQueryInlineArrayWithFloats_NotEmpty` | 446-476 | Querying an inline float array field with values returns the correct float slice. |
| `TestQueryInlineArrayWithNillableFloats` | 478-513 | Querying a nillable inline float array field returns values preserving null entries. |
| `TestQueryInlineArrayWithStrings_Null` | 515-545 | Querying an inline string array field set to null returns nil. |
| `TestQueryInlineArrayWithStrings_EmptyList` | 547-577 | Querying an inline string array field set to an empty list returns an empty slice. |
| `TestQueryInlineArrayWithStrings_NotEmpty` | 579-609 | Querying an inline string array field with values including empty string returns the correct slice. |
| `TestQueryInlineArrayWithNillableString` | 611-647 | Querying a nillable inline string array field returns values preserving null entries. |

### `with_average_filter_test.go`

AVG aggregate with element-level filters on integer and float inline arrays, including nillable variants.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArrayWithAverageWithFilter` | 21-51 | AVG aggregate on an inline integer array with a greater-than filter on elements. |
| `TestQueryInlineNillableIntegerArrayWithAverageWithFilter` | 53-83 | AVG aggregate on a nillable inline integer array with a filter excludes null values. |
| `TestQueryInlineFloatArrayWithAverageWithFilter` | 85-115 | AVG aggregate on an inline float array with a less-than filter on elements. |
| `TestQueryInlineNillableFloatArrayWithAverageWithFilter` | 117-147 | AVG aggregate on a nillable inline float array with a less-than filter excludes null values. |

### `with_average_order_test.go`

AVG aggregate used as an order key, with and without null elements in nillable arrays.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArrayWithAverageAndOrder_Succeeds` | 21-80 | AVG on inline arrays can be used as an order key in ascending and descending queries. |
| `TestQueryInlineIntegerArrayWithNullWithAverageAndOrder_Succeeds` | 82-141 | AVG on nillable inline arrays with null elements can be used as an order key. |

### `with_average_sum_test.go`

Combining AVG and SUM aggregates on the same inline integer array field.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArrayWithAverageAndSum` | 25-57 | AVG and SUM aggregates on the same inline integer array field return correct combined results. |

### `with_average_test.go`

AVG aggregate on null, empty, zero, and populated integer and float inline arrays, including nillable variants.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArrayWithAverageAndNullArray` | 21-51 | AVG on an inline integer array field set to null returns zero. |
| `TestQueryInlineIntegerArrayWithAverageAndEmptyArray` | 53-83 | AVG on an inline integer array field set to an empty array returns zero. |
| `TestQueryInlineIntegerArrayWithAverageAndZeroArray` | 85-115 | AVG on an inline integer array where all elements are zero returns zero. |
| `TestQueryInlineIntegerArrayWithAverageAndPopulatedArray` | 117-147 | AVG on a populated inline integer array returns the correct average including negatives. |
| `TestQueryInlineNillableIntegerArrayWithAverageAndPopulatedArray` | 149-179 | AVG on a nillable inline integer array ignores null values in the average calculation. |
| `TestQueryInlineFloatArrayWithAverageAndNullArray` | 181-211 | AVG on an inline float array field set to null returns zero. |
| `TestQueryInlineFloatArrayWithAverageAndEmptyArray` | 213-243 | AVG on an inline float array field set to an empty array returns zero. |
| `TestQueryInlineFloatArrayWithAverageAndZeroArray` | 245-276 | AVG on an inline float array where all elements are zero returns zero. |
| `TestQueryInlineFloatArrayWithAverageAndPopulatedArray` | 278-308 | AVG on a populated inline float array returns the correct average including negative values. |
| `TestQueryInlineNillableFloatArrayWithAverageAndPopulatedArray` | 310-340 | AVG on a nillable inline float array ignores null values in the average calculation. |

### `with_count_filter_test.go`

COUNT aggregate with element-level filters on boolean, integer, float, and string inline arrays, including nillable variants.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineBoolArrayWithCountWithFilter` | 21-51 | COUNT on an inline boolean array with an equality filter counts matching elements. |
| `TestQueryInlineNillableBoolArrayWithCountWithFilter` | 53-83 | COUNT on a nillable inline boolean array with filter counts only matching non-null elements. |
| `TestQueryInlineIntegerArrayWithCountWithFilter` | 85-115 | COUNT on an inline integer array with a greater-than filter counts matching elements. |
| `TestQueryInlineNillableIntegerArrayWithCountWithFilter` | 117-147 | COUNT on a nillable inline integer array with filter skips null elements. |
| `TestQueryInlineIntegerArrayWithsWithCountWithAndFilterAndPopulatedArray` | 149-179 | COUNT on an inline integer array with a compound _and filter counts elements matching both conditions. |
| `TestQueryInlineFloatArrayWithCountWithFilter` | 181-211 | COUNT on an inline float array with a less-than filter counts matching elements. |
| `TestQueryInlineNillableFloatArrayWithCountWithFilter` | 213-243 | COUNT on a nillable inline float array with filter counts only matching non-null elements. |
| `TestQueryInlineStringArrayWithCountWithFilter` | 245-275 | COUNT on an inline string array with an _in filter counts elements matching allowed values. |
| `TestQueryInlineNillableStringArrayWithCountWithFilter` | 277-307 | COUNT on a nillable inline string array with an _in filter skips null and non-matching elements. |

### `with_count_limit_offset_test.go`

COUNT aggregate with combined offset and limit pagination on inline integer arrays.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArrayWithCountWithOffsetWithLimitGreaterThanLength` | 21-51 | COUNT on an inline integer array with offset and limit exceeding remaining elements counts available items. |
| `TestQueryInlineIntegerArrayWithCountWithOffsetWithLimit` | 53-83 | COUNT on an inline integer array with offset and limit counts only the windowed elements. |

### `with_count_limit_test.go`

COUNT aggregate with a limit argument on inline integer arrays.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArrayWithCountWithLimitGreaterThanLength` | 21-51 | COUNT on an inline integer array with limit exceeding array length counts all available elements. |
| `TestQueryInlineIntegerArrayWithCountWithLimit` | 53-83 | COUNT on an inline integer array with limit counts only elements within the limit. |

### `with_count_order_test.go`

COUNT aggregate used as an order key on inline arrays, with and without null elements.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArray_WithCountAndOrder_Succeeds` | 21-80 | COUNT across multiple inline arrays can be used as an order key in ASC and DESC queries. |
| `TestQueryInlineIntegerArray_WithNullAndCountAndOrder_Succeeds` | 82-141 | COUNT across nillable inline arrays including null elements can be used as an order key. |

### `with_count_test.go`

COUNT aggregate on null, empty, and populated inline arrays, including nillable boolean arrays.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArrayWithCountAndNullArray` | 21-51 | COUNT on an inline integer array field set to null returns zero. |
| `TestQueryInlineIntegerArrayWithCountAndEmptyArray` | 53-83 | COUNT on an inline integer array field set to an empty list returns zero. |
| `TestQueryInlineIntegerArrayWithCountAndPopulatedArray` | 85-115 | COUNT on a populated inline integer array returns the total number of elements. |
| `TestQueryInlineNillableBoolArrayWithCountAndPopulatedArray` | 117-147 | COUNT on a nillable inline boolean array counts all elements including null entries. |

### `with_filter_all_test.go`

`_all` quantifier filter on nillable and non-null inline arrays of all supported element types, including a null-array edge case.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineStringArray_WithAllFilter_Succeeds` | 21-55 | Filtering with _all on a nillable string array returns only docs where all elements match. |
| `TestQueryInlineNotNullStringArray_WithAllFilter_Succeeds` | 57-91 | Filtering with _all on a non-null string array returns only docs where every element is non-empty. |
| `TestQueryInlineIntArray_WithAllFilter_Succeeds` | 93-127 | Filtering with _all on a nillable integer array returns only docs where no elements are null. |
| `TestQueryInlineNotNullIntArray_WithAllFilter_Succeeds` | 129-163 | Filtering with _all on a non-null integer array returns docs where all elements satisfy the condition. |
| `TestQueryInlineFloatArray_WithAllFilter_Succeeds` | 165-199 | Filtering with _all on a nillable float array returns only docs where no elements are null. |
| `TestQueryInlineNotNullFloatArray_WithAllFilter_Succeeds` | 201-235 | Filtering with _all on a non-null float array returns docs where all elements satisfy the condition. |
| `TestQueryInlineBooleanArray_WithAllFilter_Succeeds` | 237-271 | Filtering with _all on a nillable boolean array returns only docs where no elements are null. |
| `TestQueryInlineNotNullBooleanArray_WithAllFilter_Succeeds` | 273-307 | Filtering with _all on a non-null boolean array returns docs where all elements are true. |
| `TestQueryInlineStringArray_WithAllFilterAndNullValue_Succeeds` | 309-333 | Filtering with _all on a null inline string array returns no documents. |

### `with_filter_any_test.go`

`_any` quantifier filter on nillable and non-null inline arrays of all supported element types, including a null-array edge case.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineStringArray_WithAnyFilter_Succeeds` | 21-55 | Filtering with _any on a nillable string array returns docs that have at least one null element. |
| `TestQueryInlineNotNullStringArray_WithAnyFilter_Succeeds` | 57-91 | Filtering with _any on a non-null string array returns docs that contain an empty string element. |
| `TestQueryInlineIntArray_WithAnyFilter_Succeeds` | 93-127 | Filtering with _any on a nillable integer array returns docs that contain at least one null element. |
| `TestQueryInlineNotNullIntArray_WithAnyFilter_Succeeds` | 129-163 | Filtering with _any on a non-null integer array returns docs where at least one element satisfies the condition. |
| `TestQueryInlineFloatArray_WithAnyFilter_Succeeds` | 165-199 | Filtering with _any on a nillable float array returns docs that contain at least one null element. |
| `TestQueryInlineNotNullFloatArray_WithAnyFilter_Succeeds` | 201-235 | Filtering with _any on a non-null float array returns docs where at least one element satisfies the condition. |
| `TestQueryInlineBooleanArray_WithAnyFilter_Succeeds` | 237-271 | Filtering with _any on a nillable boolean array returns docs that contain at least one null element. |
| `TestQueryInlineNotNullBooleanArray_WithAnyFilter_Succeeds` | 273-307 | Filtering with _any on a non-null boolean array returns docs where any element is true. |
| `TestQueryInlineStringArray_WithAnyFilterAndNullValue_Succeeds` | 309-333 | Filtering with _any on a null inline string array returns no documents. |

### `with_filter_eq_test.go`

`_eq` and `_neq` full-array equality filters on boolean, integer, float, and string inline arrays, both non-null and nillable.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineBooleanArray_WithEqFilter_ReturnsResults` | 21-55 | Filtering with _eq on a non-null boolean array matches docs with identical array values. |
| `TestQueryInlineBooleanArray_WithNeqFilter_ReturnsResults` | 57-91 | Filtering with _neq on a non-null boolean array excludes docs with a matching array. |
| `TestQueryInlineNullableBooleanArray_WithEqFilter_ReturnsResults` | 93-127 | Filtering with _eq on a nullable boolean array matches docs with identical arrays including null. |
| `TestQueryInlineNullableBooleanArray_WithNeqFilter_ReturnsResults` | 129-163 | Filtering with _neq on a nullable boolean array excludes docs with a matching array including nulls. |
| `TestQueryInlineIntegerArray_WithEqFilter_ReturnsResults` | 165-199 | Filtering with _eq on a non-null integer array matches docs with an identical array value. |
| `TestQueryInlineIntegerArray_WithNeqFilter_ReturnsResults` | 201-235 | Filtering with _neq on a non-null integer array excludes docs with a matching array value. |
| `TestQueryInlineNullableIntegerArray_WithEqFilter_ReturnsResults` | 237-271 | Filtering with _eq on a nullable integer array matches docs with identical arrays including null. |
| `TestQueryInlineNullableIntegerArray_WithNeqFilter_ReturnsResults` | 273-307 | Filtering with _neq on a nullable integer array excludes docs with identical arrays including null. |
| `TestQueryInlineFloatArray_WithEqFilter_ReturnsResults` | 309-343 | Filtering with _eq on a non-null float array matches docs with an identical array value. |
| `TestQueryInlineFloatArray_WithNeqFilter_ReturnsResults` | 345-379 | Filtering with _neq on a non-null float array excludes docs with a matching array value. |
| `TestQueryInlineNullableFloatArray_WithEqFilter_ReturnsResults` | 380-414 | Filtering with _eq on a nullable float array matches docs with identical arrays including null. |
| `TestQueryInlineNullableFloatArray_WithNeqFilter_ReturnsResults` | 416-450 | Filtering with _neq on a nullable float array excludes docs with identical arrays including null. |
| `TestQueryInlineStringArray_WithEqFilter_ReturnsResults` | 452-486 | Filtering with _eq on a non-null string array matches docs with an identical array value. |
| `TestQueryInlineStringArray_WithNeqFilter_ReturnsResults` | 488-522 | Filtering with _neq on a non-null string array excludes docs with a matching array value. |
| `TestQueryInlineNullableStringArray_WithEqFilter_ReturnsResults` | 524-558 | Filtering with _eq on a nullable string array matches docs with identical arrays including null. |
| `TestQueryInlineNullableStringArray_WithNeqFilter_ReturnsResults` | 560-594 | Filtering with _neq on a nullable string array excludes docs with identical arrays including null. |

### `with_filter_none_test.go`

`_none` quantifier filter on nillable and non-null inline arrays of all supported element types.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineStringArrayWithNoneFilter` | 21-55 | Filtering with _none on a nillable string array returns docs where no element is null. |
| `TestQueryInlineNonNullStringArrayWithNoneFilter` | 57-91 | Filtering with _none on a non-null string array returns docs where no element is an empty string. |
| `TestQueryInlineIntArrayWithNoneFilter` | 93-127 | Filtering with _none on a nillable integer array returns docs where no element is null. |
| `TestQueryInlineNonNullIntArrayWithNoneFilter` | 129-163 | Filtering with _none on a non-null integer array returns docs where no element satisfies the condition. |
| `TestQueryInlineFloatArrayWithNoneFilter` | 165-199 | Filtering with _none on a nillable float array returns docs where no element is null. |
| `TestQueryInlineNonNullFloatArrayWithNoneFilter` | 201-235 | Filtering with _none on a non-null float array returns docs where no element satisfies the condition. |
| `TestQueryInlineBooleanArrayWithNoneFilter` | 237-271 | Filtering with _none on a nillable boolean array returns docs where no element is null. |
| `TestQueryInlineNonNullBooleanArrayWithNoneFilter` | 273-307 | Filtering with _none on a non-null boolean array returns docs where no element satisfies the condition. |

### `with_group_test.go`

`groupBy` queries using a scalar string field and an inline integer array field.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineArrayWithGroupByString` | 21-66 | Grouping by a string field returns inline array values for each document within the group. |
| `TestQueryInlineArrayWithGroupByArray` | 68-122 | Grouping by an inline integer array field groups documents with identical array values together. |

### `with_max_doc_id_test.go`

MAX aggregate on a nillable float inline array filtered by a specific docID.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineNillableFloatArray_WithDocIDAndMax_Succeeds` | 23-53 | MAX on a nillable float inline array when filtering by a specific docID returns the maximum value. |

### `with_max_filter_test.go`

MAX aggregate with element-level filters on integer and float inline arrays, including nillable variants.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArray_WithMaxWithFilter_Succeeds` | 21-51 | MAX on an inline integer array with a less-than filter returns the maximum among filtered elements. |
| `TestQueryInlineNillableIntegerArray_WithMaxWithFilter_Succeeds` | 53-83 | MAX on a nillable inline integer array with a filter ignores null and excluded elements. |
| `TestQueryInlineFloatArray_WithMaxWithFilter_Succeeds` | 85-115 | MAX on an inline float array with a less-than filter returns the maximum among filtered elements. |
| `TestQueryInlineNillableFloatArray_WithMaxWithFilter_Succeeds` | 117-147 | MAX on a nillable inline float array with a filter ignores null and excluded elements. |

### `with_max_limit_offset_order_test.go`

MAX aggregate on inline array slices defined by offset, limit, and order (ASC/DESC) for integer and float arrays.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArray_WithMaxWithOffsetWithLimitWithOrderAsc_Succeeds` | 21-52 | MAX on an inline integer array slice defined by offset, limit, and ascending order. |
| `TestQueryInlineIntegerArray_WithMaxWithOffsetWithLimitWithOrderDesc_Succeeds` | 54-85 | MAX on an inline integer array slice defined by offset, limit, and descending order. |
| `TestQueryInlineNillableIntegerArray_WithMaxWithOffsetWithLimitWithOrderAsc_Succeeds` | 87-118 | MAX on a nillable inline integer array slice defined by offset, limit, and ascending order. |
| `TestQueryInlineNillableIntegerArray_WithMaxWithOffsetWithLimitWithOrderDesc_Succeeds` | 120-151 | MAX on a nillable inline integer array slice defined by offset, limit, and descending order. |
| `TestQueryInlineFloatArray_WithMaxWithOffsetWithLimitWithOrderAsc_Succeeds` | 153-184 | MAX on an inline float array slice defined by offset, limit, and ascending order. |
| `TestQueryInlineFloatArray_WithMaxWithOffsetWithLimitWithOrderDesc_Succeeds` | 186-217 | MAX on an inline float array slice defined by offset, limit, and descending order. |
| `TestQueryInlineNillableFloatArray_WithMaxWithOffsetWithLimitWithOrderAsc_Succeeds` | 219-250 | MAX on a nillable inline float array slice defined by offset, limit, and ascending order. |
| `TestQueryInlineNillableFloatArray_WithMaxWithOffsetWithLimitWithOrderDesc_Succeeds` | 252-283 | MAX on a nillable inline float array slice defined by offset, limit, and descending order. |

### `with_max_limit_offset_test.go`

MAX aggregate on an inline integer array slice defined by offset and limit.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArray_WithMaxWithOffsetWithLimit_Succeeds` | 21-51 | MAX on an inline integer array slice defined by offset and limit returns the maximum of the window. |

### `with_max_order_test.go`

MAX aggregate used as an order key on inline arrays, with and without null elements.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArray_WithMaxAndOrder_Succeeds` | 21-80 | MAX across multiple inline arrays can be used as an order key in ASC and DESC queries. |
| `TestQueryInlineIntegerArray_WithNullAndMaxAndOrder_Succeeds` | 82-141 | MAX across nillable inline arrays with null elements can be used as an order key. |

### `with_max_test.go`

MAX aggregate on null, empty, and populated integer and float inline arrays, including nillable variants and docID filtering.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArray_WithMaxAndNullArray_Succeeds` | 21-51 | MAX on an inline integer array field set to null returns nil. |
| `TestQueryInlineIntegerArray_WithMaxAndEmptyArray_Succeeds` | 53-83 | MAX on an inline integer array field set to an empty list returns nil. |
| `TestQueryInlineIntegerArray_WithMaxAndPopulatedArray_Succeeds` | 85-115 | MAX on a populated inline integer array returns the largest element. |
| `TestQueryInlineNillableIntegerArray_WithMaxAndPopulatedArray_Succeeds` | 117-147 | MAX on a nillable inline integer array ignores null values and returns the largest element. |
| `TestQueryInlineFloatArray_WithMaxAndNullArray_Succeeds` | 149-179 | MAX on an inline float array field set to null returns nil. |
| `TestQueryInlineFloatArray_WithMaxAndEmptyArray_Succeeds` | 181-211 | MAX on an inline float array field set to an empty list returns nil. |
| `TestQueryInlineFloatArray_WithMaxAndPopulatedArray_Succeeds` | 213-243 | MAX on a populated inline float array returns the largest element. |
| `TestQueryInlineNillableFloatArray_WithMaxAndPopulatedArray_Succeeds` | 245-275 | MAX on a nillable inline float array ignores null values and returns the largest element. |
| `TestQueryInlineNillableFloatArray_WithDocIDMaxAndPopulatedArray_Succeeds` | 277-307 | MAX on a nillable inline float array filtered by a specific docID returns the correct maximum. |

### `with_min_doc_id_test.go`

MIN aggregate on a nillable float inline array filtered by a specific docID.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineNillableFloatArray_WithDocIDAndMin_Succeeds` | 23-53 | MIN on a nillable float inline array when filtering by a specific docID returns the minimum value. |

### `with_min_filter_test.go`

MIN aggregate with element-level filters on integer and float inline arrays, including nillable variants.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArray_WithMinWithFilter_Succeeds` | 21-51 | MIN on an inline integer array with a greater-than filter returns the minimum among filtered elements. |
| `TestQueryInlineNillableIntegerArray_WithMinWithFilter_Succeeds` | 53-83 | MIN on a nillable inline integer array with a filter ignores null and excluded elements. |
| `TestQueryInlineFloatArray_WithMinWithFilter_Succeeds` | 85-115 | MIN on an inline float array with a greater-than filter returns the minimum among filtered elements. |
| `TestQueryInlineNillableFloatArray_WithMinWithFilter_Succeeds` | 117-147 | MIN on a nillable inline float array with a filter ignores null and excluded elements. |

### `with_min_limit_offset_order_test.go`

MIN aggregate on inline array slices defined by offset, limit, and order (ASC/DESC) for integer and float arrays.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArray_WithMinWithOffsetWithLimitWithOrderAsc_Succeeds` | 21-52 | MIN on an inline integer array slice defined by offset, limit, and ascending order. |
| `TestQueryInlineIntegerArray_WithMinWithOffsetWithLimitWithOrderDesc_Succeeds` | 54-85 | MIN on an inline integer array slice defined by offset, limit, and descending order. |
| `TestQueryInlineNillableIntegerArray_WithMinWithOffsetWithLimitWithOrderAsc_Succeeds` | 87-118 | MIN on a nillable inline integer array slice defined by offset, limit, and ascending order. |
| `TestQueryInlineNillableIntegerArray_WithMinWithOffsetWithLimitWithOrderDesc_Succeeds` | 120-151 | MIN on a nillable inline integer array slice defined by offset, limit, and descending order. |
| `TestQueryInlineFloatArray_WithMinWithOffsetWithLimitWithOrderAsc_Succeeds` | 153-184 | MIN on an inline float array slice defined by offset, limit, and ascending order. |
| `TestQueryInlineFloatArray_WithMinWithOffsetWithLimitWithOrderDesc_Succeeds` | 186-217 | MIN on an inline float array slice defined by offset, limit, and descending order. |
| `TestQueryInlineNillableFloatArray_WithMinWithOffsetWithLimitWithOrderAsc_Succeeds` | 219-250 | MIN on a nillable inline float array slice defined by offset, limit, and ascending order. |
| `TestQueryInlineNillableFloatArray_WithMinWithOffsetWithLimitWithOrderDesc_Succeeds` | 252-283 | MIN on a nillable inline float array slice defined by offset, limit, and descending order. |

### `with_min_limit_offset_test.go`

MIN aggregate on an inline integer array slice defined by offset and limit.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArray_WithMinWithOffsetWithLimit_Succeeds` | 21-51 | MIN on an inline integer array slice defined by offset and limit returns the minimum of the window. |

### `with_min_order_test.go`

MIN aggregate used as an order key on inline arrays, with and without null elements.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArray_WithMinAndOrder_Succeeds` | 21-80 | MIN across multiple inline arrays can be used as an order key in ASC and DESC queries. |
| `TestQueryInlineIntegerArray_WithNullAndMinAndOrder_Succeeds` | 82-141 | MIN across nillable inline arrays with null elements can be used as an order key. |

### `with_min_test.go`

MIN aggregate on null, empty, and populated integer and float inline arrays, including nillable variants.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArray_WithMinAndNullArray_Succeeds` | 21-51 | MIN on an inline integer array field set to null returns nil. |
| `TestQueryInlineIntegerArray_WithMinAndEmptyArray_Succeeds` | 53-83 | MIN on an inline integer array field set to an empty list returns nil. |
| `TestQueryInlineIntegerArray_WithMinAndPopulatedArray_Succeeds` | 85-115 | MIN on a populated inline integer array returns the smallest element. |
| `TestQueryInlineNillableIntegerArray_WithMinAndPopulatedArray_Succeeds` | 117-147 | MIN on a nillable inline integer array ignores null values and returns the smallest element. |
| `TestQueryInlineFloatArray_WithMinAndNullArray_Succeeds` | 149-179 | MIN on an inline float array field set to null returns nil. |
| `TestQueryInlineFloatArray_WithMinAndEmptyArray_Succeeds` | 181-211 | MIN on an inline float array field set to an empty list returns nil. |
| `TestQueryInlineFloatArray_WithMinAndPopulatedArray_Succeeds` | 213-243 | MIN on a populated inline float array returns the smallest element. |
| `TestQueryInlineNillableFloatArray_WithMinAndPopulatedArray_Succeeds` | 245-275 | MIN on a nillable inline float array ignores null values and returns the smallest element. |

### `with_sum_filter_test.go`

SUM aggregate with element-level filters on integer and float inline arrays, including nillable variants.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArrayWithSumWithFilter` | 21-51 | SUM on an inline integer array with a greater-than filter sums only matching elements. |
| `TestQueryInlineNillableIntegerArrayWithSumWithFilter` | 53-83 | SUM on a nillable inline integer array with a filter sums only matching non-null elements. |
| `TestQueryInlineFloatArrayWithSumWithFilter` | 85-115 | SUM on an inline float array with a less-than filter sums only matching elements. |
| `TestQueryInlineNillableFloatArrayWithSumWithFilter` | 117-147 | SUM on a nillable inline float array with a filter sums only matching non-null elements. |

### `with_sum_limit_offset_order_test.go`

SUM aggregate on inline array slices defined by offset, limit, and order (ASC/DESC) for integer and float arrays.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArrayWithSumWithOffsetWithLimitWithOrderAsc` | 21-52 | SUM on an inline integer array slice defined by offset, limit, and ascending order. |
| `TestQueryInlineIntegerArrayWithSumWithOffsetWithLimitWithOrderDesc` | 54-85 | SUM on an inline integer array slice defined by offset, limit, and descending order. |
| `TestQueryInlineNillableIntegerArrayWithSumWithOffsetWithLimitWithOrderAsc` | 87-118 | SUM on a nillable inline integer array slice defined by offset, limit, and ascending order. |
| `TestQueryInlineNillableIntegerArrayWithSumWithOffsetWithLimitWithOrderDesc` | 120-151 | SUM on a nillable inline integer array slice defined by offset, limit, and descending order. |
| `TestQueryInlineFloatArrayWithSumWithOffsetWithLimitWithOrderAsc` | 153-184 | SUM on an inline float array slice defined by offset, limit, and ascending order. |
| `TestQueryInlineFloatArrayWithSumWithOffsetWithLimitWithOrderDesc` | 186-217 | SUM on an inline float array slice defined by offset, limit, and descending order. |
| `TestQueryInlineNillableFloatArrayWithSumWithOffsetWithLimitWithOrderAsc` | 219-250 | SUM on a nillable inline float array slice defined by offset, limit, and ascending order. |
| `TestQueryInlineNillableFloatArrayWithSumWithOffsetWithLimitWithOrderDesc` | 252-283 | SUM on a nillable inline float array slice defined by offset, limit, and descending order. |

### `with_sum_limit_offset_test.go`

SUM aggregate on an inline integer array slice defined by offset and limit.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArrayWithSumWithOffsetWithLimit` | 21-51 | SUM on an inline integer array slice defined by offset and limit sums only the windowed elements. |

### `with_sum_limit_test.go`

SUM aggregate with a limit argument on an inline integer array.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArrayWithSumWithLimit` | 21-51 | SUM on an inline integer array with a limit sums only the first N elements. |

### `with_sum_order_test.go`

SUM aggregate used as an order key on inline arrays, with and without null elements.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArray_WithSumAndOrder_Succeeds` | 21-81 | SUM across multiple inline arrays can be used as an order key in ASC and DESC queries. |
| `TestQueryInlineIntegerArray_WithNullAndSumAndOrder_Succeeds` | 83-143 | SUM across nillable inline arrays with null elements can be used as an order key. |

### `with_sum_test.go`

SUM aggregate on null, empty, and populated integer and float inline arrays, including nillable variants.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryInlineIntegerArrayWithSumAndNullArray` | 21-51 | SUM on an inline integer array field set to null returns zero. |
| `TestQueryInlineIntegerArrayWithSumAndEmptyArray` | 53-83 | SUM on an inline integer array field set to an empty list returns zero. |
| `TestQueryInlineIntegerArrayWithSumAndPopulatedArray` | 85-115 | SUM on a populated inline integer array returns the sum of all elements including negatives. |
| `TestQueryInlineNillableIntegerArrayWithSumAndPopulatedArray` | 117-147 | SUM on a nillable inline integer array ignores null values and sums the remaining elements. |
| `TestQueryInlineFloatArrayWithSumAndNullArray` | 149-179 | SUM on an inline float array field set to null returns zero. |
| `TestQueryInlineFloatArrayWithSumAndEmptyArray` | 181-211 | SUM on an inline float array field set to an empty list returns zero. |
| `TestQueryInlineFloatArrayWithSumAndPopulatedArray` | 213-243 | SUM on a populated inline float array returns the sum of all float elements. |
| `TestQueryInlineNillableFloatArrayWithSumAndPopulatedArray` | 245-275 | SUM on a nillable inline float array ignores null values and sums the remaining elements. |
