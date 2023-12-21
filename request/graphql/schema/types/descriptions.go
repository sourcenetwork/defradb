// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package types

const (
	OrderArgDescription string = `
An optional set of field-orders which may be used to sort the results. An
 empty set will be ignored.
`
	GroupByArgDescription string = `
An optional set of fields for which to group the contents of this field by.
 If this argument is provided, only fields used to group may be rendered in
 the immediate child selector.  Additional fields may be selected by using
 the '_group' selector within the immediate child selector. If an empty set
 is provided, the restrictions mentioned still apply, although all results
 will appear within the same group.
`
	LimitArgDescription string = `
An optional value that caps the number of results to the number provided.
 A limit of zero will be ignored.
`
	OffsetArgDescription string = `
An optional value that skips the given number of results that would have
 otherwise been returned.  Commonly used alongside the 'limit' argument,
 this argument will still work on its own.
`
	commitDescription string = `
Commit represents an individual commit to a MerkleCRDT, every mutation to a
 document will result in a new commit per modified field, and one composite
 commit composed of the field level commits and, in the case of an update,
 the prior composite commit.
`
	commitDocIDArgDescription string = `
An optional docID parameter for this commit query. Only commits for a document
 with a matching docID will be returned.  If no documents match, the result
 set will be empty.
`
	commitFieldIDArgDescription string = `
An optional field ID parameter for this commit query. Only commits for a fields
 matching this ID will be returned. Specifying 'C' will limit the results to 
 composite (document level) commits only, otherwise field IDs are numeric. If no
 fields match, the result set will be empty.
`
	commitCIDArgDescription string = `
An optional value that specifies the commit ID of the commits to return. If a
 matching commit is not found then an empty set will be returned.
`
	commitDepthArgDescription string = `
An optional value that specifies the maximum depth to which the commit DAG graph
 should be traversed from matching commits.
`
	commitLinksDescription string = `
Child commits in the DAG that contribute to the composition of this commit.
 Composite commits will link to the field commits for the fields modified during
 the single mutation.
`
	commitHeightFieldDescription string = `
Height represents the location of the commit in the DAG. All commits (composite,
 and field level) on create will have a height of '1', each subsequent local update
 will increment this by one for the new commits.
`
	commitCIDFieldDescription string = `
The unique CID of this commit, and the primary means through which to safely identify
 a specific commit.
`
	commitDocIDFieldDescription string = `
The docID of the document that this commit is for.
`
	commitCollectionIDFieldDescription string = `
The ID of the collection that this commit was committed against.
`
	commitSchemaVersionIDFieldDescription string = `
The ID of the schema version that this commit was committed against. This ID allows one
 to determine the state of the data model at the time of commit.
`
	commitFieldNameFieldDescription string = `
The name of the field that this commit was committed against. If this is a composite field
 the value will be null.
`
	commitFieldIDFieldDescription string = `
The id of the field that this commit was committed against. If this is a composite field
 the value will be "C".
`
	commitDeltaFieldDescription string = `
The CBOR encoded representation of the value that is saved as part of this commit.
`
	commitLinkNameFieldDescription string = `
The Name of the field that this linked commit mutated.
`
	commitLinkCIDFieldDescription string = `
The CID of this linked commit.
`
	commitFieldsEnumDescription string = `
These are the set of fields supported for grouping by in a commits query.
`
	commitsQueryDescription string = `
Returns a set of commits matching any provided criteria. If no arguments are
 provided all commits in the system will be returned.
`
	latestCommitsQueryDescription string = `
Returns a set of head commits matching any provided criteria. If no arguments are
 provided all head commits in the system will be returned. If no 'field' argument
 is provided only composite commits will be returned. This is equivalent to
 a 'commits' query with Depth: 1, and a differing 'field' default value.
`
	CountFieldDescription string = `
Returns the total number of items within the specified child sets. If multiple child
 sets are specified, the combined total of all of them will be returned as a single value.
`
	SumFieldDescription string = `
Returns the total sum of the specified field values within the specified child sets. If
 multiple fields/sets are specified, the combined sum of all of them will be returned as
 a single value.
`
	AverageFieldDescription string = `
Returns the average of the specified field values within the specified child sets. If
 multiple fields/sets are specified, the combined average of all items within each set
 (true average, not an average of averages) will be returned as a single value.
`
	booleanOperatorBlockDescription string = `
These are the set of filter operators available for use when filtering on Boolean
 values.
`
	notNullBooleanOperatorBlockDescription string = `
These are the set of filter operators available for use when filtering on Boolean!
 values.
`
	dateTimeOperatorBlockDescription string = `
These are the set of filter operators available for use when filtering on DateTime
 values.
`
	floatOperatorBlockDescription string = `
These are the set of filter operators available for use when filtering on Float
 values.
`
	notNullFloatOperatorBlockDescription string = `
These are the set of filter operators available for use when filtering on Float!
 values.
`
	intOperatorBlockDescription string = `
These are the set of filter operators available for use when filtering on Int
 values.
`
	notNullIntOperatorBlockDescription string = `
These are the set of filter operators available for use when filtering on Int!
 values.
`
	stringOperatorBlockDescription string = `
These are the set of filter operators available for use when filtering on String
 values.
`
	notNullStringOperatorBlockDescription string = `
These are the set of filter operators available for use when filtering on String!
 values.
`
	idOperatorBlockDescription string = `
These are the set of filter operators available for use when filtering on ID
 values.
`
	eqOperatorDescription string = `
The equality operator - if the target matches the value the check will pass.
`
	neOperatorDescription string = `
The inequality operator - if the target does not matches the value the check will pass.
`
	inOperatorDescription string = `
The contains operator - if the target value is within the given set the check will pass.
`
	ninOperatorDescription string = `
The does not contains operator - if the target value is not within the given set the
 check will pass.
`
	gtOperatorDescription string = `
The greater than operator - if the target value is greater than the given value the
 check will pass.
`
	geOperatorDescription string = `
The greater than or equal to operator - if the target value is greater than or equal to the
 given value the check will pass.
`
	ltOperatorDescription string = `
The less than operator - if the target value is less than the given value the check will pass.
`
	leOperatorDescription string = `
The less than or equal to operator - if the target value is less than or equal to the
 given value the check will pass.
`
	likeStringOperatorDescription string = `
The like operator - if the target value contains the given sub-string the check will pass. '%'
 characters may be used as wildcards, for example '_like: "%Ritchie"' would match on strings
 ending in 'Ritchie'.
`
	nlikeStringOperatorDescription string = `
The not-like operator - if the target value does not contain the given sub-string the check will
 pass. '%' characters may be used as wildcards, for example '_nlike: "%Ritchie"' would match on
 the string 'Quentin Tarantino'.
`
	AndOperatorDescription string = `
The and operator - all checks within this clause must pass in order for this check to pass.
`
	OrOperatorDescription string = `
The or operator - only one check within this clause must pass in order for this check to pass.
`
	NotOperatorDescription string = `
The negative operator - this check will only pass if all checks within it fail.
`
	ascOrderDescription string = `
Sort the results in ascending order, e.g. null,1,2,3,a,b,c.
`
	descOrderDescription string = `
Sort the results in descending order, e.g. c,b,a,3,2,1,null.
`
	primaryDirectiveDescription string = `
Indicate the primary side of a one-to-one relationship.
`
	relationDirectiveDescription string = `
Allows the explicit definition of relationship attributes instead of using the system generated
 defaults.
`
	relationDirectiveNameArgDescription string = `
Explicitly define the name of the relationship instead of using the system generated defaults.
`
)
