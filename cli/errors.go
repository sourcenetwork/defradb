// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"fmt"

	"github.com/sourcenetwork/defradb/errors"
)

const (
	errInvalidLensConfig            string = "invalid lens configuration"
	errSchemaVersionNotOfSchema     string = "the given schema version is from a different schema"
	errRequiredFlag                 string = "the required flag [--%s|-%s] is %s"
	errInvalidAscensionOrder        string = "invalid order: expected ASC or DESC"
	errInvalidInxedFieldDescription string = "invalid or malformed field description"
)

var (
	ErrNoDocOrFile                = errors.New("document or file must be defined")
	ErrInvalidDocument            = errors.New("invalid document")
	ErrNoDocIDOrFilter            = errors.New("docID or filter must be defined")
	ErrInvalidExportFormat        = errors.New("invalid export format")
	ErrNoLensConfig               = errors.New("lens config cannot be empty")
	ErrInvalidLensConfig          = errors.New("invalid lens configuration")
	ErrSchemaVersionNotOfSchema   = errors.New(errSchemaVersionNotOfSchema)
	ErrViewAddMissingArgs         = errors.New("please provide a base query and output SDL for this view")
	ErrPolicyFileArgCanNotBeEmpty = errors.New("policy file argument can not be empty")
	ErrPurgeForceFlagRequired     = errors.New("run this command again with --force if you really want to purge all data")
	ErrMissingKeyringSecret       = errors.New("missing keyring secret")
)

func NewErrRequiredFlagEmpty(longName string, shortName string) error {
	return errors.New(fmt.Sprintf(errRequiredFlag, longName, shortName, "empty"))
}

func NewErrRequiredFlagInvalid(longName string, shortName string) error {
	return errors.New(fmt.Sprintf(errRequiredFlag, longName, shortName, "invalid"))
}

func NewErrInvalidLensConfig(inner error) error {
	return errors.Wrap(errInvalidLensConfig, inner)
}

func NewErrSchemaVersionNotOfSchema(schemaRoot string, schemaVersionID string) error {
	return errors.New(
		errSchemaVersionNotOfSchema,
		errors.NewKV("SchemaRoot", schemaRoot),
		errors.NewKV("SchemaVersionID", schemaVersionID),
	)
}

func NewErrInvalidAscensionOrder(fieldName string) error {
	return errors.New(errInvalidAscensionOrder, errors.NewKV("Field", fieldName))
}

func NewErrInvalidInxedFieldDescription(fieldName string) error {
	return errors.New(errInvalidInxedFieldDescription, errors.NewKV("Field", fieldName))
}
