// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package document

import "github.com/sourcenetwork/defradb/client"

// Field is an interface to interact with Fields inside a document
type Field interface {
	Name() string
	Type() client.CType //TODO Abstract into a Field Type interface
	SchemaType() string
}

type simpleField struct {
	name       string
	crdtType   client.CType
	schemaType string
}

func (doc *Document) newField(t client.CType, name string, schemaType ...string) Field {
	f := simpleField{
		name:     name,
		crdtType: t,
	}
	if len(schemaType) > 0 {
		f.schemaType = schemaType[0]
	}
	return f
}

func (field simpleField) Name() string {
	return field.name
}

func (field simpleField) Type() client.CType {
	return field.crdtType
}

func (field simpleField) SchemaType() string {
	return field.schemaType
}
