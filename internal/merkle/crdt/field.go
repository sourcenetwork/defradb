package merklecrdt

import "github.com/sourcenetwork/defradb/client"

// DocField is a struct that holds the document ID and the field value.
// This is used to a link between the document and the field value.
// For example, to check if the field value need be encrypted depending on the document-level
// encryption is enabled or not.
type DocField struct {
	DocID      string
	FieldValue *client.FieldValue
}
