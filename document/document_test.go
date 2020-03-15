package document

import (
	"testing"

	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/document/key"
	"github.com/sourcenetwork/defradb/merkle/crdt"
)

var (
	testJSONObj = []byte(`{
		"Name": "John",
		"Age": 26,
		"Address": {
			"Street": "Main",
			"City": "Toronto"
		}
	}`)

	pref = cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.SHA2_256,
		MhLength: -1, // default length
	}
)

func TestNewFromJSON(t *testing.T) {
	doc, err := NewFromJSON(testJSONObj)
	if err != nil {
		t.Error("Error creating new doc from JSON:", err)
		return
	}

	c, err := pref.Sum(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	objKey := key.NewDocKeyV0(c)
	if objKey.String() != doc.Key().String() {
		t.Errorf("Incorrect doc key. Want %v, have %v", objKey.String(), doc.Key().String())
		return
	}

	// check field/value
	// fields
	assert.Equal(t, doc.fields["Name"].Name(), "Name")
	assert.Equal(t, doc.fields["Name"].Type(), crdt.LWW_REGISTER)
	assert.Equal(t, doc.fields["Age"].Name(), "Age")
	assert.Equal(t, doc.fields["Age"].Type(), crdt.LWW_REGISTER)
	assert.Equal(t, doc.fields["Address"].Name(), "Address")
	assert.Equal(t, doc.fields["Address"].Type(), crdt.OBJECT)

	//values
	assert.Equal(t, doc.values[doc.fields["Name"]].Value(), "John")
	assert.Equal(t, doc.values[doc.fields["Name"]].IsDocument(), false)
	assert.Equal(t, doc.values[doc.fields["Age"]].Value(), float64(26))
	assert.Equal(t, doc.values[doc.fields["Age"]].IsDocument(), false)
	assert.Equal(t, doc.values[doc.fields["Address"]].IsDocument(), true)

	//subdoc fields
	subDoc := doc.values[doc.fields["Address"]].Value().(*Document)
	assert.Equal(t, subDoc.fields["Street"].Name(), "Street")
	assert.Equal(t, subDoc.fields["Street"].Type(), crdt.LWW_REGISTER)
	assert.Equal(t, subDoc.fields["City"].Name(), "City")
	assert.Equal(t, subDoc.fields["City"].Type(), crdt.LWW_REGISTER)

	//subdoc values
	assert.Equal(t, subDoc.values[subDoc.fields["Street"]].Value(), "Main")
	assert.Equal(t, subDoc.values[subDoc.fields["Street"]].IsDocument(), false)
	assert.Equal(t, subDoc.values[subDoc.fields["City"]].Value(), "Toronto")
}
