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

type DocsList struct {
	ColName string
	Docs    []map[string]any
}

type GeneratedDoc struct {
	ColIndex int
	JSON     string
}

type tStr = string
type fStr = string

type fieldDefinition struct {
	name       fStr
	typeStr    tStr
	isArray    bool
	isRelation bool
	isPrimary  bool
}

type typeDefinition struct {
	name   tStr
	index  int
	fields []fieldDefinition
}

type genConfig struct {
	labels []string
	props  map[string]any
}

func (t *typeDefinition) getField(name fStr) *fieldDefinition {
	for i := range t.fields {
		if t.fields[i].name == name {
			return &t.fields[i]
		}
	}
	return nil
}
