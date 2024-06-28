// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package mapper

// Field contains the most basic information about a requestable.
type Field struct {
	// The location of this field within it's parent.
	Index int

	// The name of this field.  For example 'Age', or '_group'.
	Name string
}

func (f *Field) GetIndex() int {
	return f.Index
}

func (f *Field) GetName() string {
	return f.Name
}

func (f *Field) AsTargetable() (*Targetable, bool) {
	return nil, false
}

func (f *Field) AsSelect() (*Select, bool) {
	return nil, false
}

func (f *Field) CloneTo(index int) Requestable {
	return f.cloneTo(index)
}

func (f *Field) cloneTo(index int) *Field {
	return &Field{
		Index: index,
		Name:  f.Name,
	}
}
