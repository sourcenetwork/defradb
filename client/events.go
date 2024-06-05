package client

import "github.com/ipfs/go-cid"

const (
	// UpdateEventName is the name of the database update event.
	UpdateEventName = "db:update"
	// ResultsEventName is the name of the database results event.
	ResultsEventName = "db:results"
)

// UpdateEvent represents a new DAG node added to the append-only composite MerkleCRDT Clock graph
// of a document.
//
// It must only contain public elements not protected by ACP.
type UpdateEvent struct {
	// DocID is the unique immutable identifier of the document that was updated.
	DocID string

	// Cid is the id of the composite commit that formed this update in the DAG.
	Cid cid.Cid

	// SchemaRoot is the root identifier of the schema that defined the shape of the document that was updated.
	SchemaRoot string

	// Block is the encoded contents of this composite commit, it contains the Cids of the field level commits that
	// also formed this update.
	Block []byte

	// IsCreate is true if this update is the creation of a new document.
	IsCreate bool
}
