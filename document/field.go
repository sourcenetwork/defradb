// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package document

import (
	"github.com/sourcenetwork/defradb/core"
)

// Field is an interface to interact with Fields inside a document
type Field interface {
	Key() core.DataStoreKey
	Name() string
	Type() core.CType //TODO Abstract into a Field Type interface
	SchemaType() string
}

type simpleField struct {
	name       string
	key        core.DataStoreKey
	crdtType   core.CType
	schemaType string
}

func (doc *Document) newField(t core.CType, name string, schemaType ...string) Field {
	f := simpleField{
		name:     name,
		key:      doc.Key().Key.WithFieldId(name),
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

func (field simpleField) Type() core.CType {
	return field.crdtType
}

func (field simpleField) Key() core.DataStoreKey {
	return field.key
}

func (field simpleField) SchemaType() string {
	return field.schemaType
}
