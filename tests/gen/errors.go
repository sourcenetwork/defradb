// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

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
