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

import (
	"fmt"

	"github.com/sourcenetwork/defradb/errors"
)

type protocolError string

func (e protocolError) Error() string        { return string(e) }
func (e protocolError) Is(target error) bool { return target == ErrGraphQLProtocol }

func newErrExpectedNull(expectedType string) error {
	return protocolError(fmt.Sprintf("Expected %q, found null.", expectedType))
}

var (
	ErrFilterMissingArgumentType      = errors.New("couldn't find filter argument type")
	ErrInvalidOrderDirection          = errors.New("invalid order direction")
	ErrInvalidOrderInput              = errors.New("invalid order input")
	ErrFailedToParseConditionsFromAST = errors.New("couldn't parse conditions value from AST")
	ErrFailedToParseConditionValue    = errors.New("failed to parse condition value from query filter statement")
	ErrEmptyDataPayload               = errors.New("given data payload is empty")
	ErrUnknownMutationName            = errors.New("unknown mutation name")
	ErrTruncateFilterNull             = errors.New("truncate filter cannot be null")
	ErrInvalidExplainTypeArg          = errors.New("invalid explain request type argument")
	ErrInvalidNumberOfExplainArgs     = errors.New("invalid number of arguments to an explain request")
	ErrUnknownExplainType             = errors.New("invalid / unknown explain type")
	ErrUnknownGQLOperation            = errors.New("unknown GraphQL operation type")
	ErrInvalidFilterConditions        = errors.New("invalid filter condition type, expected map")
	ErrMultipleOrderFieldsDefined     = errors.New("each order argument can only define one field")
	ErrMultipleDocIDsNotSupported     = errors.New("querying by multiple docIDs is not yet supported")
	ErrSimilarityMissingTarget        = errors.New("similarity requires a target field argument")
	ErrGraphQLProtocol                = errors.New("graphql protocol error")
)

const errSimilarityOnNonVectorField string = "similarity can only target a numeric array field"

// NewErrSimilarityOnNonVectorField returns an error indicating that similarity was given a field
// that cannot hold a vector.
func NewErrSimilarityOnNonVectorField(fieldName string, fieldType string) error {
	return errors.New(
		errSimilarityOnNonVectorField,
		errors.NewKV("Field", fieldName),
		errors.NewKV("Type", fieldType),
	)
}
