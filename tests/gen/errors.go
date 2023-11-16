// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package gen

import "github.com/sourcenetwork/defradb/errors"

const (
	errInvalidConfiguration    string = "invalid configuration"
	errCanNotSupplyTypeDemand  string = "can not supply demand for type "
	errFailedToParse           string = "failed to parse schema"
	errFailedToGenerateDoc     string = "failed to generate doc"
	errIncompleteColDefinition string = "incomplete collection definition"
)

func NewErrInvalidConfiguration(reason string) error {
	return errors.New(errInvalidConfiguration, errors.NewKV("Reason", reason))
}

func NewErrCanNotSupplyTypeDemand(typeName string) error {
	return errors.New(errCanNotSupplyTypeDemand, errors.NewKV("Type", typeName))
}

func NewErrFailedToParse(reason string) error {
	return errors.New(errFailedToParse, errors.NewKV("Reason", reason))
}

func NewErrFailedToGenerateDoc(inner error) error {
	return errors.Wrap(errFailedToGenerateDoc, inner)
}

func NewErrIncompleteColDefinition(reason string) error {
	return errors.New(errIncompleteColDefinition, errors.NewKV("Reason", reason))
}

func newNotDefinedTypeErr(typeName string) error {
	return NewErrInvalidConfiguration("type " + typeName + " is not defined in the schema")
}
