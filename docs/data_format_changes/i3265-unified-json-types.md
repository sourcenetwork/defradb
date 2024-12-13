# Unified JSON Types

Applied a common interface to all JSON types which made it use float64 for all numbers. 
This in turned caused encoded data to change because CBOR encoding of float64 is different from int64.
