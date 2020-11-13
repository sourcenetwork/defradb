package base

type DataEncoding uint32

const (
	// Indicates that the data is encoded using the CBOR encoding for values.
	DataEncoding_VALUE_CBOR DataEncoding = 0
)
