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

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ccid "github.com/sourcenetwork/defradb/core/cid"
)

var (
	testJSONObj = []byte(`{
		"Name": "John",
		"Age": 26
	}`)

	pref = ccid.NewDefaultSHA256PrefixV1()

	schemaDescriptions = []SchemaDescription{
		{
			Name: "User",
			Fields: []FieldDescription{
				{
					Name: "Name",
					Typ:  LWW_REGISTER,
					Kind: FieldKind_NILLABLE_STRING,
				},
				{
					Name: "Age",
					Typ:  LWW_REGISTER,
					Kind: FieldKind_NILLABLE_INT,
				},
			},
		},
	}
)

func TestNewFromJSON(t *testing.T) {
	doc, err := NewDocFromJSON(testJSONObj, schemaDescriptions[0])
	if err != nil {
		t.Error("Error creating new doc from JSON:", err)
		return
	}

	buf, err := doc.Bytes()
	if err != nil {
		t.Error(err)
	}

	// And then feed it some data
	c, err := pref.Sum(buf)
	if err != nil {
		t.Error(err)
	}
	objKey := NewDocIDV0(c)

	if objKey.String() != doc.ID().String() {
		t.Errorf("Incorrect document ID. Want %v, have %v", objKey.String(), doc.ID().String())
		return
	}

	// check field/value
	// fields
	assert.Equal(t, doc.fields["Name"].Name(), "Name")
	assert.Equal(t, doc.fields["Name"].Type(), LWW_REGISTER)
	assert.Equal(t, doc.fields["Age"].Name(), "Age")
	assert.Equal(t, doc.fields["Age"].Type(), LWW_REGISTER)

	//values
	assert.Equal(t, doc.values[doc.fields["Name"]].Value(), "John")
	assert.Equal(t, doc.values[doc.fields["Name"]].IsDocument(), false)
	assert.Equal(t, doc.values[doc.fields["Age"]].Value(), int64(26))
	assert.Equal(t, doc.values[doc.fields["Age"]].IsDocument(), false)
}

func TestSetWithJSON(t *testing.T) {
	doc, err := NewDocFromJSON(testJSONObj, schemaDescriptions[0])
	if err != nil {
		t.Error("Error creating new doc from JSON:", err)
		return
	}

	buf, err := doc.Bytes()
	if err != nil {
		t.Error(err)
	}

	// And then feed it some data
	c, err := pref.Sum(buf)
	if err != nil {
		t.Error(err)
	}
	objKey := NewDocIDV0(c)

	if objKey.String() != doc.ID().String() {
		t.Errorf("Incorrect document ID. Want %v, have %v", objKey.String(), doc.ID().String())
		return
	}

	updatePatch := []byte(`{
		"Name": "Alice",
		"Age": 27
	}`)
	err = doc.SetWithJSON(updatePatch)
	if err != nil {
		t.Error(err)
	}

	// check field/value
	// fields
	assert.Equal(t, doc.fields["Name"].Name(), "Name")
	assert.Equal(t, doc.fields["Name"].Type(), LWW_REGISTER)
	assert.Equal(t, doc.fields["Age"].Name(), "Age")
	assert.Equal(t, doc.fields["Age"].Type(), LWW_REGISTER)

	//values
	assert.Equal(t, doc.values[doc.fields["Name"]].Value(), "Alice")
	assert.Equal(t, doc.values[doc.fields["Name"]].IsDocument(), false)
	assert.Equal(t, doc.values[doc.fields["Age"]].Value(), int64(27))
	assert.Equal(t, doc.values[doc.fields["Age"]].IsDocument(), false)
}

func TestNewDocsFromJSON_WithObjectInsteadOfArray_Error(t *testing.T) {
	_, err := NewDocsFromJSON(testJSONObj, schemaDescriptions[0])
	require.ErrorContains(t, err, "value doesn't contain array; it contains object")
}
