// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package planner

import "github.com/sourcenetwork/defradb/errors"

const (
	errUnknownDependency string = "The given field does not exist"
)

var (
	ErrDeltaMissingPriority                = errors.New("commit Delta missing priority key")
	ErrFailedToFindScanNode                = errors.New("failed to find original scan node in plan graph")
	ErrMissingQueryOrMutation              = errors.New("query is missing query or mutation statements")
	ErrOperationDefinitionMissingSelection = errors.New("operationDefinition is missing selections")
	ErrFailedToFindGroupSource             = errors.New("failed to identify group source")
	ErrGroupOutsideOfGroupBy               = errors.New("_group may only be referenced when within a groupBy query")
	ErrMissingChildSelect                  = errors.New("expected child select but none was found")
	ErrMissingChildValue                   = errors.New("expected child value, however none was yielded")
	ErrUnknownRelationType                 = errors.New("failed sub selection, unknown relation type")
	ErrUnknownDependency                   = errors.New(errUnknownDependency)
)

func NewErrUnknownDependency(name string) error {
	return errors.New(errUnknownDependency, errors.NewKV("Name", name))
}
