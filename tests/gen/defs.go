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

const (
	stringType = "String"
	intType    = "Int"
	boolType   = "Boolean"
	floatType  = "Float"
)

type DocsList struct {
	ColName string
	Docs    []map[string]any
}

type GeneratedDoc struct {
	ColIndex int
	JSON     string
}

type fieldDefinition struct {
	name       string
	typeStr    string
	isArray    bool
	isRelation bool
	isPrimary  bool
}

type typeDefinition struct {
	name   string
	index  int
	fields []fieldDefinition
}

type GenerateFieldFunc func(i int, next func() any) any

type genConfig struct {
	labels         []string
	props          map[string]any
	fieldGenerator GenerateFieldFunc
}

func (t *typeDefinition) getField(name string) *fieldDefinition {
	for i := range t.fields {
		if t.fields[i].name == name {
			return &t.fields[i]
		}
	}
	return nil
}
