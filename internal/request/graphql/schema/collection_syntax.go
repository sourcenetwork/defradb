// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.

package schema

type collectionDocument struct {
	Definitions []*typeDefinition
}

type collectionName struct {
	Value string
}

type collectionFieldDefinition struct {
	Name       *collectionName
	Type       collectionType
	Directives []*collectionDirective
}

type collectionDirective struct {
	Name      *collectionName
	Arguments []*collectionArgument
}

type collectionArgument struct {
	Name  *collectionName
	Value collectionValue
}

type collectionType interface {
	String() string
	isCollectionType()
}

type collectionNamed struct{ Name *collectionName }
type collectionList struct{ Type collectionType }
type collectionNonNull struct{ Type collectionType }

func (*collectionNamed) isCollectionType()   {}
func (*collectionList) isCollectionType()    {}
func (*collectionNonNull) isCollectionType() {}
func (*collectionNamed) String() string      { return "Named" }
func (*collectionList) String() string       { return "List" }
func (*collectionNonNull) String() string    { return "NonNull" }

type collectionValue interface {
	GetValue() any
	isCollectionValue()
}

type collectionVariable struct{ Name *collectionName }
type collectionIntValue struct{ Value string }
type collectionFloatValue struct{ Value string }
type collectionStringValue struct{ Value string }
type collectionBooleanValue struct{ Value bool }
type collectionNullValue struct{}
type collectionEnumValue struct{ Value string }
type collectionListValue struct{ Values []collectionValue }
type collectionObjectValue struct{ Fields []*collectionObjectField }
type collectionObjectField struct {
	Name  *collectionName
	Value collectionValue
}

func (*collectionVariable) isCollectionValue()     {}
func (*collectionIntValue) isCollectionValue()     {}
func (*collectionFloatValue) isCollectionValue()   {}
func (*collectionStringValue) isCollectionValue()  {}
func (*collectionBooleanValue) isCollectionValue() {}
func (*collectionNullValue) isCollectionValue()    {}
func (*collectionEnumValue) isCollectionValue()    {}
func (*collectionListValue) isCollectionValue()    {}
func (*collectionObjectValue) isCollectionValue()  {}

func (v *collectionVariable) GetValue() any     { return v.Name }
func (v *collectionIntValue) GetValue() any     { return v.Value }
func (v *collectionFloatValue) GetValue() any   { return v.Value }
func (v *collectionStringValue) GetValue() any  { return v.Value }
func (v *collectionBooleanValue) GetValue() any { return v.Value }
func (*collectionNullValue) GetValue() any      { return nil }
func (v *collectionEnumValue) GetValue() any    { return v.Value }
func (v *collectionListValue) GetValue() any    { return v.Values }
func (v *collectionObjectValue) GetValue() any  { return v.Fields }
