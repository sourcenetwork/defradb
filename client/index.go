package client

type IndexDirection string

const (
	Ascending  IndexDirection = "ASC"
	Descending IndexDirection = "DESC"
)

type IndexedFieldDescription struct {
	Name      string
	Direction IndexDirection
}

type IndexDescription struct {
	Name   string
	ID     uint32
	Fields []IndexedFieldDescription
	Unique bool
}

type CollectionIndexDescription struct {
	CollectionName string
	Index          IndexDescription
}
