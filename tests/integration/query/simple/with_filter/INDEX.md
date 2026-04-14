# Index: `tests/integration/query/simple/with_filter`

## Overview

This folder contains integration tests for the `filter` argument on simple (non-relational) collection queries. The tests cover all comparison operators (`_eq`, `_neq`, `_gt`, `_geq`, `_lt`, `_leq`, `_in`, `_nin`, `_like`, `_ilike`, `_nlike`, `_nilike`) across every supported scalar field type (String, Int, Float, DateTime, Boolean, Blob), as well as logical combinators (`_and`, `_or`, `_not`) and alias-based filtering. Null-value edge cases and inline array field variants are also exercised.

## Test Index

### `with_alias_test.go`

Tests filtering using the `_alias` filter key, covering aliased field names, empty/null/non-object alias values, non-existent aliases, and compound alias expressions.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithAliasEqualsFilterBlock_ShouldFilter` | 21-57 | Filter by _alias using an aliased field name with _eq operator returns matching document. |
| `TestQuerySimple_WithEmptyAlias_ShouldNotFilter` | 59-100 | An empty _alias filter object applies no filtering and returns all documents. |
| `TestQuerySimple_WithNullAlias_ShouldFilterAll` | 102-133 | A null _alias filter value filters out all documents and returns an empty result. |
| `TestQuerySimple_WithNonObjectAlias_ShouldFilterAll` | 135-166 | A non-object _alias filter value filters out all documents and returns an empty result. |
| `TestQuerySimple_WithNonExistantAlias_ShouldReturnError` | 168-197 | Filtering by an alias name that does not exist in the select returns an error. |
| `TestQuerySimple_WithNonAliasedField_ShouldMatchFilter` | 199-235 | Filtering by _alias using the original field name without an alias still applies the filter. |
| `TestQuerySimple_WithCompoundAlias_ShouldMatchFilter` | 237-278 | Compound _and filter using an aliased field name correctly narrows the result set. |
| `TestQuerySimple_WithAliasWithCompound_ShouldMatchFilter` | 280-323 | _alias filter containing a nested _and compound expression correctly filters using aliased field. |

### `with_and_test.go`

Tests the `_and` logical combinator on scalar and inline-array fields, combining comparison operators across one or more conditions.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithIntGreaterThanAndIntLessThanFilter` | 21-74 | _and filter combining _gt and _lt on an integer field returns only matching documents. |
| `TestQuerySimple_WithInlineIntArray_GreaterThanAndLessThanFilter_Succeeds` | 76-119 | _and filter on an inline int array field using _all _geq and _lt returns the matching document. |

### `with_eq_blob_test.go`

Tests `_eq` filtering on a Blob field.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithEqOpOnBlobField_ShouldFilter` | 21-61 | Filter by _eq on a Blob field returns only the document with the matching hex value. |

### `with_eq_datetime_test.go`

Tests `_eq` filtering on a DateTime field, including exact match, null match, and the case where some documents have nil DateTime values.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithDateTimeEqualsFilterBlock` | 21-61 | Filter by _eq on a DateTime field returns only the document with the matching timestamp. |
| `TestQuerySimpleWithDateTimeEqualsNilFilterBlock` | 63-109 | Filter by _eq null on a DateTime field returns only the document with no datetime set. |
| `TestQuerySimple_WithNilDateTimeEqualsAndNonNilFilterBlock_ShouldSucceed` | 111-157 | Filter by _eq on a DateTime field skips the nil-DateTime document and returns the exact match. |

### `with_eq_float_test.go`

Tests `_eq` filtering on a Float field, including exact match and null match.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithFloatEqualsFilterBlock` | 21-57 | Filter by _eq on a Float field returns only the document with the exact matching value. |
| `TestQuerySimpleWithFloatEqualsNilFilterBlock` | 59-100 | Filter by _eq null on a Float field returns only the document with no float value set. |

### `with_eq_int_test.go`

Tests `_eq` filtering on an integer field, including exact match and null match.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithIntEqualsFilterBlock` | 21-57 | Filter by _eq on an integer field returns only the document with the exact matching value. |
| `TestQuerySimpleWithIntEqualsNilFilterBlock` | 59-100 | Filter by _eq null on an integer field returns only the document with no integer value set. |

### `with_eq_string_test.go`

Tests `_eq` filtering on a string field, including null match, and verifies correct field selection behaviour when the filtered field differs from the selected fields.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithStringFilterBlock` | 21-57 | Filter by _eq on a string field returns only the document with the exact matching string. |
| `TestQuerySimpleWithStringEqualsNilFilterBlock` | 59-100 | Filter by _eq null on a string field returns only the document with no string value set. |
| `TestQuerySimpleWithStringFilterBlockAndSelect_SelectSameFieldAsFilterWithMatch` | 102-136 | Selecting the same string field used in an _eq filter returns only the matching document. |
| `TestQuerySimpleWithStringFilterBlockAndSelect_SelectDifferentFieldThanFilterWithMatch` | 137-171 | Selecting a field different from the _eq string filter field returns the filtered document's other field. |
| `TestQuerySimpleWithStringFilterBlockAndSelect_SelectMultipleFieldsButNoMatch` | 172-197 | Filtering by a string value that matches no document returns an empty result set. |

### `with_geq_datetime_test.go`

Tests `_geq` filtering on a DateTime field across equal, greater, lesser, null, and mixed-null scenarios.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithDateTimeGEFilterBlockWithEqualValue` | 21-57 | _geq filter on a DateTime field with the boundary value returns the boundary document. |
| `TestQuerySimpleWithDateTimeGEFilterBlockWithGreaterValue` | 59-95 | _geq filter on a DateTime field with a value less than the document's date returns that document. |
| `TestQuerySimpleWithDateTimeGEFilterBlockWithLesserValue` | 97-129 | _geq filter on a DateTime field with a value after all documents returns no results. |
| `TestQuerySimpleWithDateTimeGEFilterBlockWithNilValue` | 131-168 | _geq null filter on a DateTime field returns all documents regardless of whether the field is set. |
| `TestQuerySimple_WithNilDateTimeGEAndNonNilFilterBlock_ShouldSucceed` | 170-222 | _geq filter on a DateTime field skips the nil-DateTime document and returns the matching ones. |

### `with_geq_float_test.go`

Tests `_geq` filtering on a Float field, including equal boundary, slightly-below boundary, integer threshold, and null.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithHeightMGEFilterBlockWithEqualValue` | 21-55 | _geq filter on a Float field with the boundary value returns the boundary document. |
| `TestQuerySimpleWithHeightMGEFilterBlockWithLesserValue` | 57-91 | _geq filter on a Float field with a value just below the boundary still matches the boundary document. |
| `TestQuerySimpleWithHeightMGEFilterBlockWithLesserIntValue` | 93-127 | _geq filter on a Float field with an integer threshold returns documents with float values at or above it. |
| `TestQuerySimpleWithHeightMGEFilterBlockWithNilValue` | 129-166 | _geq null filter on a Float field returns all documents including those with no float value set. |

### `with_geq_int_test.go`

Tests `_geq` filtering on an integer field, including equal boundary, below-boundary, and null threshold.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithIntGEFilterBlockWithEqualValue` | 21-55 | _geq filter on an integer field with the boundary value returns the boundary document. |
| `TestQuerySimpleWithIntGEFilterBlockWithGreaterValue` | 57-91 | _geq filter on an integer field with a value below the target returns the greater document. |
| `TestQuerySimpleWithIntGEFilterBlockWithNilValue` | 93-130 | _geq null filter on an integer field returns all documents including those with no integer value set. |

### `with_gt_datetime_test.go`

Tests `_gt` filtering on a DateTime field across strictly-greater, one-day-before, after-all, null, and mixed-null scenarios.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithDateTimeGTFilterBlockWithEqualValue` | 21-57 | _gt filter on a DateTime field returns the document with a date strictly greater than the threshold. |
| `TestQuerySimpleWithDateTimeGTFilterBlockWithGreaterValue` | 59-95 | _gt filter on a DateTime field with a value one day before the document's date returns that document. |
| `TestQuerySimpleWithDateTimeGTFilterBlockWithLesserValue` | 97-129 | _gt filter on a DateTime field with a threshold after all documents returns an empty result. |
| `TestQuerySimpleWithDateTimeGTFilterBlockWithNilValue` | 131-164 | _gt null filter on a DateTime field returns only documents that have a non-null datetime value. |
| `TestQuerySimple_WithNilDateTimeGTAndNonNilFilterBlock_ShouldSucceed` | 166-212 | _gt filter on a DateTime field skips the nil-DateTime document and returns the strictly greater one. |

### `with_gt_float_test.go`

Tests `_gt` filtering on a Float field, covering one match, no match, all match, integer threshold, and null threshold.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithFloatGreaterThanFilterBlock_OneMatchingResult` | 21-55 | _gt filter on a Float field returns one matching document that strictly exceeds the threshold. |
| `TestQuerySimpleWithFloatGreaterThanFilterBlock_NoMatchingResult` | 57-87 | _gt filter on a Float field with a threshold greater than all documents returns an empty result. |
| `TestQuerySimpleWithFloatGreaterThanFilterBlock_AllMatchingResult` | 89-127 | _gt filter on a Float field with a threshold below all documents returns all documents. |
| `TestQuerySimpleWithFloatGreaterThanFilterBlockWithIntFilterValue` | 129-163 | _gt filter on a Float field using an integer threshold correctly returns documents above it. |
| `TestQuerySimpleWithFloatGreaterThanFilterBlockWithNullFilterValue` | 165-198 | _gt null filter on a Float field returns only documents that have a non-null float value. |

### `with_gt_int_test.go`

Tests `_gt` filtering on an integer field, covering one match, no match, multiple matches, and null threshold.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithIntGreaterThanFilterBlock_ReturnOneAsOneMatches` | 21-57 | _gt filter on an integer field returns the one document that strictly exceeds the threshold. |
| `TestQuerySimpleWithIntGreaterThanFilterBlock_ReturnNoneAsNoMatch` | 59-90 | _gt filter on an integer field with a threshold above all documents returns an empty result. |
| `TestQuerySimpleWithIntGreaterThanFilterBlock_ReturnAllMultiMatches` | 92-133 | _gt filter on an integer field with a low threshold returns all documents that exceed it. |
| `TestQuerySimpleWithIntGreaterThanFilterBlockWithNullFilterValue` | 135-168 | _gt null filter on an integer field returns only documents that have a non-null integer value. |

### `with_in_blob_test.go`

Tests `_in` filtering on a Blob field.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithInOpOnBlobField_ShouldFilter` | 21-61 | _in filter on a Blob field returns only the document whose hex value is in the provided list. |

### `with_in_test.go`

Tests `_in` filtering on integer and float fields, including a null value in the list.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithIntInFilter` | 21-74 | _in filter on an integer field returns only documents whose age is in the provided value list. |
| `TestQuerySimpleWithIntInFilterOnFloat` | 76-125 | _in filter on a Float field using mixed integer and float values returns documents with exact matches. |
| `TestQuerySimpleWithIntInFilterWithNullValue` | 127-189 | _in filter including null in the list returns documents with nil integer field alongside matching values. |

### `with_leq_datetime_test.go`

Tests `_leq` filtering on a DateTime field across equal boundary, greater-than-boundary, null, and mixed-null scenarios.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithDateTimeLEFilterBlockWithEqualValue` | 21-57 | _leq filter on a DateTime field with the boundary value returns the document at the boundary. |
| `TestQuerySimpleWithDateTimeLEFilterBlockWithGreaterValue` | 59-95 | _leq filter on a DateTime field with a threshold after the document's date returns that document. |
| `TestQuerySimpleWithDateTimeLEFilterBlockWithNullValue` | 97-132 | _leq null filter on a DateTime field returns only documents that have no datetime value set. |
| `TestQuerySimple_WithNilDateTimeLEAndNonNilFilterBlock_ShouldSucceed` | 134-186 | _leq filter on a DateTime field skips the nil-DateTime document and returns those at or before the threshold. |

### `with_leq_float_test.go`

Tests `_leq` filtering on a Float field, covering equal boundary, just-above boundary, integer threshold, and null.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithFloatLEFilterBlockWithEqualValue` | 21-55 | _leq filter on a Float field with the boundary value returns the document at the boundary. |
| `TestQuerySimpleWithFloatLEFilterBlockWithGreaterValue` | 57-91 | _leq filter on a Float field with a threshold just above the boundary returns the lower document. |
| `TestQuerySimpleWithFloatLEFilterBlockWithGreaterIntValue` | 93-127 | _leq filter on a Float field using an integer threshold returns documents at or below that integer. |
| `TestQuerySimpleWithFloatLEFilterBlockWithNullValue` | 129-162 | _leq null filter on a Float field returns only documents that have no float value set. |

### `with_leq_int_test.go`

Tests `_leq` filtering on an integer field, covering equal boundary, above-boundary, and null threshold.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithIntLEFilterBlockWithEqualValue` | 21-55 | _leq filter on an integer field with the boundary value returns the document at the boundary. |
| `TestQuerySimpleWithIntLEFilterBlockWithGreaterValue` | 57-91 | _leq filter on an integer field with a threshold above the target returns the lower document. |
| `TestQuerySimpleWithIntLEFilterBlockWithNullValue` | 93-126 | _leq null filter on an integer field returns only documents that have no integer value set. |

### `with_like_blob_test.go`

Tests `_like` pattern filtering on a Blob field.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithLikeOpOnBlobField_ShouldFilter` | 21-61 | _like filter on a Blob field with a wildcard pattern returns documents whose hex value matches. |

### `with_like_string_test.go`

Tests `_like` and `_ilike` pattern filtering on a string field, covering substring, prefix, suffix, exact, start-and-end, compound logical, and nil-field cases.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithLikeStringContainsFilterBlockContainsString` | 21-55 | _like filter with a substring wildcard pattern returns only the document whose name contains the substring. |
| `TestQuerySimple_WithCaseInsensitiveLike_ShouldMatchString` | 57-91 | _ilike filter with a lowercase substring pattern returns the document regardless of original casing. |
| `TestQuerySimpleWithLikeStringContainsFilterBlockAsPrefixString` | 93-127 | _like filter with a prefix pattern returns only the document whose name starts with that prefix. |
| `TestQuerySimple_WithCaseInsensitiveLikeString_ShouldMatchPrefixString` | 129-163 | _ilike filter with a lowercase prefix pattern returns the document whose name starts with that prefix. |
| `TestQuerySimpleWithLikeStringContainsFilterBlockAsSuffixString` | 165-199 | _like filter with a suffix pattern returns only the document whose name ends with that suffix. |
| `TestQuerySimple_WithCaseInsensitiveLikeString_ShouldMatchSuffixString` | 201-235 | _ilike filter with a lowercase suffix pattern returns the document whose name ends with that suffix. |
| `TestQuerySimpleWithLikeStringContainsFilterBlockExactString` | 237-271 | _like filter with an exact string and no wildcards returns only the document with that exact name. |
| `TestQuerySimple_WithCaseInsensitiveLikeString_ShouldMatchExactString` | 273-307 | _ilike filter with a full lowercase string matches the document with a different-case exact name. |
| `TestQuerySimpleWithLikeStringContainsFilterBlockContainsStringMuplitpleResults` | 309-347 | _like filter with a common substring pattern returns all documents whose names contain that substring. |
| `TestQuerySimpleWithLikeStringContainsFilterBlockHasStartAndEnd` | 349-383 | _like filter with a prefix and suffix wildcard pattern returns the document matching both ends. |
| `TestQuerySimpleWithLikeStringContainsFilterBlockHasBoth` | 385-415 | _and of two _like conditions requiring two substrings returns no results when no document matches both. |
| `TestQuerySimpleWithLikeStringContainsFilterBlockHasEither` | 417-451 | _or of two _like conditions returns documents matching either of the two substring patterns. |
| `TestQuerySimpleWithLikeStringContainsFilterBlockPropNotSet` | 453-492 | _like filter does not match documents whose Name field is nil when the pattern requires a value. |

### `with_lt_datetime_test.go`

Tests `_lt` filtering on a DateTime field, covering strictly-less-than, null threshold, and mixed-null scenarios.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithDateTimeLTFilterBlockWithGreaterValue` | 21-57 | _lt filter on a DateTime field returns the document whose date is strictly before the threshold. |
| `TestQuerySimpleWithDateTimeLTFilterBlockWithNullValue` | 59-90 | _lt null filter on a DateTime field returns no documents because nothing is less than null. |
| `TestQuerySimple_WithNilDateTimeLTAndNonNilFilterBlock_ShouldSucceed` | 92-139 | _lt filter on a DateTime field skips the nil-DateTime document and returns the strictly lesser one. |

### `with_lt_float_test.go`

Tests `_lt` filtering on a Float field, covering strictly-less-than with a float and integer threshold, and null threshold.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithFloatLessThanFilterBlockWithGreaterValue` | 21-55 | _lt filter on a Float field returns the document whose value is strictly less than the threshold. |
| `TestQuerySimpleWithFloatLessThanFilterBlockWithGreaterIntValue` | 57-91 | _lt filter on a Float field using an integer threshold returns the document below that integer. |
| `TestQuerySimpleWithFloatLessThanFilterBlockWithNullValue` | 93-123 | _lt null filter on a Float field returns no documents because nothing is less than null. |

### `with_lt_int_test.go`

Tests `_lt` filtering on an integer field, covering strictly-less-than and null threshold.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithIntLessThanFilterBlockWithGreaterValue` | 21-55 | _lt filter on an integer field returns the document whose value is strictly less than the threshold. |
| `TestQuerySimpleWithIntLessThanFilterBlockWithNullValue` | 57-86 | _lt null filter on an integer field returns no documents because nothing is less than null. |

### `with_neq_bool_test.go`

Tests `_neq` filtering on a Boolean field, covering not-true, not-null, and not-false cases.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithBoolNotEqualsTrueFilterBlock` | 21-64 | _neq true filter on a Boolean field returns documents that are false or have no value set. |
| `TestQuerySimpleWithBoolNotEqualsNilFilterBlock` | 66-108 | _neq null filter on a Boolean field returns only documents with a non-null boolean value. |
| `TestQuerySimpleWithBoolNotEqualsFalseFilterBlock` | 110-153 | _neq false filter on a Boolean field returns documents that are true or have no value set. |

### `with_neq_datetime_test.go`

Tests `_neq` filtering on a DateTime field, including not-matching a specific timestamp, not-null, and mixed-null documents.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithDateTimeNotEqualsFilterBlock` | 21-57 | _neq filter on a DateTime field returns only documents whose date does not match the given timestamp. |
| `TestQuerySimpleWithDateTimeNotEqualsNilFilterBlock` | 59-105 | _neq null filter on a DateTime field returns only documents that have a non-null datetime value. |
| `TestQuerySimple_WithNilDateTimeNotEqualAndNonNilFilterBlock_ShouldSucceed` | 107-159 | _neq filter on a DateTime field returns nil-DateTime and mismatched-value documents. |

### `with_neq_float_test.go`

Tests `_neq` filtering on a Float field, including not-matching a specific value and not-null.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithFloatNotEqualsFilterBlock` | 21-55 | _neq filter on a Float field returns only documents whose value does not equal the given float. |
| `TestQuerySimpleWithFloatNotEqualsNilFilterBlock` | 57-100 | _neq null filter on a Float field returns only documents that have a non-null float value. |

### `with_neq_int_test.go`

Tests `_neq` filtering on an integer field, including not-matching a specific value and not-null.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithIntNotEqualsFilterBlock` | 21-55 | _neq filter on an integer field returns only documents whose value does not equal the given integer. |
| `TestQuerySimpleWithIntNotEqualsNilFilterBlock` | 57-100 | _neq null filter on an integer field returns only documents that have a non-null integer value. |

### `with_neq_string_test.go`

Tests `_neq` filtering on a string field, including not-matching a specific value and not-null.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithStringNotEqualsFilterBlock` | 21-55 | _neq filter on a string field returns only documents whose value does not equal the given string. |
| `TestQuerySimpleWithStringNotEqualsNilFilterBlock` | 57-100 | _neq null filter on a string field returns only documents that have a non-null string value. |

### `with_nin_test.go`

Tests `_nin` filtering on an integer field, excluding a list that includes null values.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithNotInFilter` | 21-76 | _nin filter on an integer field excluding null returns only documents not in the specified list. |

### `with_nlike_string_test.go`

Tests `_nlike` and `_nilike` (negated pattern) filtering on a string field, covering substring, prefix, suffix, exact, start-and-end, compound logical, and nil-field cases.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithNotLikeStringContainsFilterBlockContainsString` | 21-55 | _nlike filter with a substring pattern returns only documents whose name does not contain that substring. |
| `TestQuerySimple_WithNotCaseInsensitiveLikeString_ShouldMatchString` | 57-91 | _nilike filter with a lowercase substring excludes matching documents regardless of original casing. |
| `TestQuerySimpleWithNotLikeStringContainsFilterBlockAsPrefixString` | 93-127 | _nlike filter with a prefix pattern returns only documents whose name does not start with that prefix. |
| `TestQuerySimple_WithNotCaseInsensitiveLikeString_ShouldMatchPrefixString` | 129-163 | _nilike filter with a lowercase prefix excludes documents starting with that prefix regardless of casing. |
| `TestQuerySimpleWithNotLikeStringContainsFilterBlockAsSuffixString` | 165-199 | _nlike filter with a suffix pattern returns only documents whose name does not end with that suffix. |
| `TestQuerySimple_WithNotCaseInsensitiveLikeString_ShouldMatchSuffixString` | 201-235 | _nilike filter with a lowercase suffix excludes documents ending with that suffix regardless of casing. |
| `TestQuerySimpleWithNotLikeStringContainsFilterBlockExactString` | 237-271 | _nlike filter with an exact string excludes the document matching that name exactly. |
| `TestQuerySimple_WithNotCaseInsensitiveLikeString_MatchExactString` | 273-307 | _nilike filter with a full lowercase string excludes the document with that exact name regardless of casing. |
| `TestQuerySimpleWithNotLikeStringContainsFilterBlockContainsStringMuplitpleResults` | 309-339 | _nlike filter with a common substring excludes all documents containing that substring and returns none. |
| `TestQuerySimpleWithNotLikeStringContainsFilterBlockHasStartAndEnd` | 341-375 | _nlike filter with a prefix and suffix pattern returns documents not matching both ends simultaneously. |
| `TestQuerySimpleWithNotLikeStringContainsFilterBlockHasBoth` | 377-411 | _and of two _nlike conditions returns documents not matching either substring pattern. |
| `TestQuerySimpleWithNotLikeStringContainsFilterBlockHasEither` | 413-451 | _or of two _nlike conditions returns all documents that do not match at least one of the patterns. |
| `TestQuerySimpleWithNotLikeStringContainsFilterBlockPropNotSet` | 453-495 | _nlike filter includes nil-Name documents along with documents that do not match the pattern. |

### `with_not_test.go`

Tests the `_not` logical filter, including wrapping equality, comparison, and compound `_or` conditions, as well as the empty `_not` and nested `_not` edge cases.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithNotEqualToXFilter_NoError` | 21-78 | _not filter with an _eq condition excludes documents matching that value and returns the rest. |
| `TestQuerySimple_WithNotAndComparisonXFilter_NoError` | 80-128 | _not filter with a _gt comparison returns only documents that do not exceed the threshold. |
| `TestQuerySimple_WithNotEqualToXorYFilter_NoError` | 130-183 | _not filter wrapping an _or condition excludes documents matching either of the two conditions. |
| `TestQuerySimple_WithEmptyNotFilter_ReturnError` | 185-228 | An empty _not filter object returns an empty result set with no error. |
| `TestQuerySimple_WithNotEqualToXAndNotYFilter_NoError` | 230-297 | Nested _not filters combining _eq and _not conditions returns documents not satisfying both exclusions simultaneously. |

### `with_or_test.go`

Tests the `_or` logical combinator on scalar and inline-array fields.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithIntEqualToXOrYFilter` | 21-74 | _or filter with two _eq conditions on an integer field returns documents matching either value. |
| `TestQuerySimple_WithInlineIntArray_EqualToXOrYFilter_Succeeds` | 76-123 | _or filter on an inline int array field using _any operator returns all documents with any qualifying element. |
