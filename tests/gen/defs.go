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

type propDefinition struct {
	name       string
	typeStr    string
	isArray    bool
	isRelation bool
	isPrimary  bool
}

type typeDefinition struct {
	name  string
	index int
	props []propDefinition
}

func (t *typeDefinition) getProp(name string) *propDefinition {
	for i := range t.props {
		if t.props[i].name == name {
			return &t.props[i]
		}
	}
	return nil
}
