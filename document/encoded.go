package document

import (
	"github.com/sourcenetwork/defradb/db/base"
)

type EncProperty struct {
	Desc base.FieldDescription
	Raw  []byte
}

// @todo: Implement Encoded Document type
type EncodedDocument struct {
	Key        []byte
	Schema     *base.SchemaDescription
	Properties map[string]*EncProperty
}

func (encdoc *EncodedDocument) Reset() {
	encdoc.Properties = make(map[string]*EncProperty)
	encdoc.Key = nil
}
