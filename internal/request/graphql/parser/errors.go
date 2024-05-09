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

var (
	ErrFilterMissingArgumentType      = errors.New("couldn't find filter argument type")
	ErrInvalidOrderDirection          = errors.New("invalid order direction string")
	ErrFailedToParseConditionsFromAST = errors.New("couldn't parse conditions value from AST")
	ErrFailedToParseConditionValue    = errors.New("failed to parse condition value from query filter statement")
	ErrEmptyDataPayload               = errors.New("given data payload is empty")
	ErrUnknownMutationName            = errors.New("unknown mutation name")
	ErrInvalidExplainTypeArg          = errors.New("invalid explain request type argument")
	ErrInvalidNumberOfExplainArgs     = errors.New("invalid number of arguments to an explain request")
	ErrUnknownExplainType             = errors.New("invalid / unknown explain type")
	ErrUnknownGQLOperation            = errors.New("unknown GraphQL operation type")
	ErrInvalidFilterConditions        = errors.New("invalid filter condition type, expected map")
)
