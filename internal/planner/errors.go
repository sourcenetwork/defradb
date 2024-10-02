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
	errUnknownDependency              string = "given field does not exist"
	errFailedToClosePlan              string = "failed to close the plan"
	errFailedToCollectExecExplainInfo string = "failed to collect execution explain information"
	errSubTypeInit                    string = "sub-type initialization error at scan node reset"
)

var (
	ErrDeltaMissingSchemaVersionID         = errors.New("commit Delta missing schema version ID")
	ErrDeltaMissingPriority                = errors.New("commit Delta missing priority key")
	ErrDeltaMissingDocID                   = errors.New("commit Delta missing document ID")
	ErrDeltaMissingFieldName               = errors.New("commit Delta missing field name")
	ErrFailedToFindScanNode                = errors.New("failed to find original scan node in plan graph")
	ErrMissingQueryOrMutation              = errors.New("request is missing query or mutation operation statements")
	ErrOperationDefinitionMissingSelection = errors.New("operationDefinition is missing selections")
	ErrFailedToFindGroupSource             = errors.New("failed to identify group source")
	ErrCantExplainSubscriptionRequest      = errors.New("can not explain a subscription request")
	ErrGroupOutsideOfGroupBy               = errors.New("_group may only be referenced when within a groupBy request")
	ErrMissingChildSelect                  = errors.New("expected child select but none was found")
	ErrMissingChildValue                   = errors.New("expected child value, however none was yielded")
	ErrUnknownRelationType                 = errors.New("failed sub selection, unknown relation type")
	ErrUnknownExplainRequestType           = errors.New("can not explain request of unknown type")
	ErrUpsertMultipleDocuments             = errors.New("cannot upsert multiple matching documents")
)

func NewErrUnknownDependency(name string) error {
	return errors.New(errUnknownDependency, errors.NewKV("Name", name))
}

func NewErrFailedToClosePlan(inner error, location string) error {
	return errors.Wrap(errFailedToClosePlan, inner, errors.NewKV("Location", location))
}

func NewErrFailedToCollectExecExplainInfo(inner error) error {
	return errors.Wrap(errFailedToCollectExecExplainInfo, inner)
}

func NewErrSubTypeInit(inner error) error {
	return errors.Wrap(errSubTypeInit, inner)
}
