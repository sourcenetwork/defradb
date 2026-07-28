// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package parser

import "github.com/sourcenetwork/defradb/errors"

const (
	errInvalidBm25Field   string = "invalid _bm25 field entry"
	errInvalidBm25Boost   string = "invalid _bm25 field weight"
	errDuplicateBm25Field string = "_bm25 may name each field only once"
)

var (
	ErrBm25NoFields = errors.New("_bm25 requires at least one field to score")
	ErrFilterMissingArgumentType      = errors.New("couldn't find filter argument type")
	ErrInvalidOrderDirection          = errors.New("invalid order direction")
	ErrInvalidOrderInput              = errors.New("invalid order input")
	ErrFailedToParseConditionsFromAST = errors.New("couldn't parse conditions value from AST")
	ErrFailedToParseConditionValue    = errors.New("failed to parse condition value from query filter statement")
	ErrEmptyDataPayload               = errors.New("given data payload is empty")
	ErrUnknownMutationName            = errors.New("unknown mutation name")
	ErrInvalidExplainTypeArg          = errors.New("invalid explain request type argument")
	ErrInvalidNumberOfExplainArgs     = errors.New("invalid number of arguments to an explain request")
	ErrUnknownExplainType             = errors.New("invalid / unknown explain type")
	ErrUnknownGQLOperation            = errors.New("unknown GraphQL operation type")
	ErrInvalidFilterConditions        = errors.New("invalid filter condition type, expected map")
	ErrMultipleOrderFieldsDefined     = errors.New("each order argument can only define one field")
	ErrMultipleDocIDsNotSupported     = errors.New("querying by multiple docIDs is not yet supported")
)

// NewErrInvalidBm25Field returns an error for an entry of the _bm25 fields argument that names no
// field.
func NewErrInvalidBm25Field(entry any) error {
	return errors.New(errInvalidBm25Field, errors.NewKV("Entry", entry))
}

// NewErrInvalidBm25Boost returns an error for an entry of the _bm25 fields argument whose weight
// is not a number zero or greater.
func NewErrInvalidBm25Boost(entry string, weight string) error {
	return errors.New(errInvalidBm25Boost, errors.NewKV("Entry", entry), errors.NewKV("Weight", weight))
}

// NewErrDuplicateBm25Field returns an error for a field named more than once in the _bm25 fields
// argument.
func NewErrDuplicateBm25Field(name string) error {
	return errors.New(errDuplicateBm25Field, errors.NewKV("Field", name))
}
