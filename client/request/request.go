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
	Directives Directives
	Selections []Selection
}
