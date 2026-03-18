package client

// CollectionFieldMapping contains the short ID mapping for a single collection,
// enabling cross-instance KV import by mapping source short IDs to FieldID CIDs.
type CollectionFieldMapping struct {
	// CollectionID is the content-addressed CID of the collection.
	CollectionID string `json:"collection_id"`
	// CollectionShortID is the source instance's short collection ID.
	CollectionShortID uint32 `json:"collection_short_id"`
	// FieldIDMapping maps source short field ID → FieldID (CID string).
	FieldIDMapping map[uint32]string `json:"field_id_mapping"`
}
