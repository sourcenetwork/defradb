# Change encoding from protobuf to cbor and use the new IPLD schema

The DAG blocks are now encoded using CBOR instead of protobuf and we use the new `github.com/ipld/go-ipld-prime` package to handle block encoding and decoding. It makes use of the new IPLD schema to define the block structure.