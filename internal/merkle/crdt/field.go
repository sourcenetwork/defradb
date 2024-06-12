package merklecrdt

import "github.com/sourcenetwork/defradb/client"

type Field struct {
	DocID      string
	FieldValue *client.FieldValue
}
