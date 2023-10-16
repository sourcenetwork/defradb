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
	for _, prop := range t.props {
		if prop.name == name {
			return &prop
		}
	}
	return nil
}
