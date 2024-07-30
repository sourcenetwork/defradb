// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package request

import (
	"github.com/sourcenetwork/immutable"
)

type Request struct {
	Queries      []*OperationDefinition
	Mutations    []*OperationDefinition
	Subscription []*OperationDefinition
}

// Operation returns the operation that should be executed from the request.
//
// The operationName is used to select an operation when multiple are defined.
func (r *Request) Operation(operationName string) (*OperationDefinition, error) {
	switch {
	case len(r.Queries) == 1 && len(r.Mutations) == 0 && len(r.Subscription) == 0:
		return r.Queries[0], nil

	case len(r.Queries) == 0 && len(r.Mutations) == 1 && len(r.Subscription) == 0:
		return r.Mutations[0], nil

	case len(r.Queries) == 0 && len(r.Mutations) == 0 && len(r.Subscription) == 1:
		return r.Subscription[0], nil

	case len(r.Queries) == 0 && len(r.Mutations) == 0 && len(r.Subscription) == 0:
		return nil, ErrMissingQueryOrMutation
	}
	for _, op := range r.Queries {
		if op.Name == operationName {
			return op, nil
		}
	}
	for _, op := range r.Mutations {
		if op.Name == operationName {
			return op, nil
		}
	}
	for _, op := range r.Subscription {
		if op.Name == operationName {
			return op, nil
		}
	}
	return nil, ErrMissingOperationName
}

type Selection any

// Directives contains all the optional and non-optional
// directives (and their additional data) that a request can have.
//
// An optional directive has a value if it's found in the request.
type Directives struct {
	// ExplainType is an optional directive (`@explain`) and it's type information.
	ExplainType immutable.Option[ExplainType]
}

type OperationDefinition struct {
	Name       string
	Selections []Selection
	Directives Directives
}
