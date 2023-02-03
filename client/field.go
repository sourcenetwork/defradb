// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

// Field is an interface to interact with Fields inside a document.
type Field interface {
	Name() string
	Type() CType //TODO Abstract into a Field Type interface
	SchemaType() string
}

type simpleField struct {
	name       string
	crdtType   CType
	schemaType string
}

func (doc *Document) newField(t CType, name string, schemaType ...string) Field {
	f := simpleField{
		name:     name,
		crdtType: t,
	}
	if len(schemaType) > 0 {
		f.schemaType = schemaType[0]
	}
	return f
}

// Name returns the name of the field.
func (field simpleField) Name() string {
	return field.name
}

// Type returns the type of the field.
func (field simpleField) Type() CType {
	return field.crdtType
}

// SchemaType returns the schema type of the field.
func (field simpleField) SchemaType() string {
	return field.schemaType
}
