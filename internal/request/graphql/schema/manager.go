// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"io"

	wgast "github.com/wundergraph/graphql-go-tools/v2/pkg/ast"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astnormalization"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astparser"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astprinter"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astvalidation"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/operationreport"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
)

// SchemaManager creates an instanced management point
// for schema intake/outtake, and updates.
type SchemaManager struct {
	definition                    *wgast.Document
	isSearchableEncryptionEnabled bool
}

// NewSchemaManager returns a manager initialized with the canonical default
// schema definition.
func NewSchemaManager(isSearchableEncryptionEnabled bool) (*SchemaManager, error) {
	sm := &SchemaManager{isSearchableEncryptionEnabled: isSearchableEncryptionEnabled}
	if err := sm.setDefinition(defaultSchemaSDL); err != nil {
		return nil, err
	}
	return sm, nil
}

// Definition returns the normalized graphql-go-tools schema snapshot. The
// document is immutable after the manager is published and must not be mutated
// by callers.
func (s *SchemaManager) Definition() *wgast.Document {
	return s.definition
}

func (s *SchemaManager) setDefinition(sdl string) error {
	document, report := astparser.ParseGraphqlDocumentString(sdl)
	if report.HasErrors() {
		return report
	}
	astnormalization.NormalizeDefinition(&document, &report)
	if report.HasErrors() {
		return report
	}
	validationReport := operationreport.Report{}
	if astvalidation.DefaultDefinitionValidator().Validate(&document, &validationReport) == astvalidation.Invalid {
		return validationReport
	}
	s.definition = &document
	return nil
}

func (s *SchemaManager) ParseSDL(sdl string) ([]core.Collection, error) {
	document, report := astparser.ParseGraphqlDocumentString(sdl)
	if report.HasErrors() {
		return nil, report
	}
	collectionDocument := adaptCollectionDocument(&document)
	return fromAst(collectionDocument)
}

func (s *SchemaManager) WriteSDL(writer io.Writer) error {
	err := astprinter.PrintIndent(s.definition, []byte("    "), writer)
	if err != nil {
		return errors.Join(ErrWritingSDL, err)
	}
	return nil
}
