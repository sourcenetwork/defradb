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

type Request struct {
	Queries      []*OperationDefinition
	Mutations    []*OperationDefinition
	Subscription []*OperationDefinition
}

type Selection any

type OperationDefinition struct {
	Selections []Selection
	IsExplain  bool
}
