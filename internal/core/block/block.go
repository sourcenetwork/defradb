// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package coreblock

import (
	"bytes"
	"sort"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/multiformats/go-multicodec"

	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
)

// Schema is the IPLD schema type that represents a `Block`.
var (
	Schema                    schema.Type
	SchemaPrototype           ipld.NodePrototype
	EncryptionSchema          schema.Type
	EncryptionSchemaPrototype ipld.NodePrototype
)

func init() {
	Schema, SchemaPrototype = mustSetSchema(
		"Block",
		&Block{},
		&DAGLink{},
		&crdt.CRDT{},
		&crdt.LWWRegDelta{},
		&crdt.CompositeDAGDelta{},
		&crdt.CounterDelta{},
	)

	EncryptionSchema, EncryptionSchemaPrototype = mustSetSchema(
		"Encryption",
		&Encryption{},
	)
}

type schemaDefinition interface {
	// IPLDSchemaBytes returns the IPLD schema representation for the type.
	IPLDSchemaBytes() []byte
}

func mustSetSchema(schemaName string, schemas ...schemaDefinition) (schema.Type, ipld.NodePrototype) {
	schemaBytes := make([][]byte, 0, len(schemas))
	for _, s := range schemas {
		schemaBytes = append(schemaBytes, s.IPLDSchemaBytes())
	}

	ts, err := ipld.LoadSchemaBytes(bytes.Join(schemaBytes, nil))
	if err != nil {
		panic(err)
	}
	blockSchemaType := ts.TypeByName(schemaName)

	// Calling bindnode.Prototype here ensure that [Block] and all the types it contains
	// are compatible with the IPLD schema defined by blockSchemaType.
	// If [Block] and `blockSchematype` do not match, this will panic.
	proto := bindnode.Prototype(schemas[0], blockSchemaType)

	return blockSchemaType, proto.Representation()
}

// DAGLink represents a link to another object in a DAG.
type DAGLink struct {
	// Name is the name of the link.
	//
	// This will be either the field name of the CRDT delta or "_head" for the head link.
	Name string
	// Link is the CID link to the object.
	cidlink.Link
}

// IPLDSchemaBytes returns the IPLD schema representation for the DAGLink.
//
// This needs to match the [DAGLink] struct or [mustSetSchema] will panic on init.
func (l DAGLink) IPLDSchemaBytes() []byte {
	return []byte(`
	type DAGLink struct { 
		name	String
		link 	Link
	}`)
}

func NewDAGLink(name string, link cidlink.Link) DAGLink {
	return DAGLink{
		Name: name,
		Link: link,
	}
}

// Encryption contains the encryption information for the block's delta.
type Encryption struct {
	// DocID is the ID of the document that is encrypted with the associated encryption key.
	DocID []byte
	// FieldName is the name of the field that is encrypted with the associated encryption key.
	// It is set if encryption is applied to a field instead of the whole doc.
	// It needs to be a pointer so that it can be translated from and to `optional` in the IPLD schema.
	FieldName *string
	// Encryption key.
	Key []byte
}

// Block is a block that contains a CRDT delta and links to other blocks.
type Block struct {
	// Delta is the CRDT delta that is stored in the block.
	Delta crdt.CRDT

	// The previous block-CIDs that this block is based on.
	//
	// For example:
	// - This will be empty for all 'create' blocks.
	// - Most 'update' blocks will have a single head, however they will have multiple if the history has
	//   diverged and there were multiple blocks at the previous height.
	Heads []cidlink.Link

	// Links are the links to other blocks in the DAG.
	//
	// This does not include `Heads`.  This will be empty for field-level blocks. It will never be empty
	// for composite blocks (and will contain links to field-level blocks).
	Links []DAGLink

	// Encryption contains the encryption information for the block's delta.
	// It needs to be a pointer so that it can be translated from and to `optional` in the IPLD schema.
	Encryption *cidlink.Link
}

// IsEncrypted returns true if the block is encrypted.
func (block *Block) IsEncrypted() bool {
	return block.Encryption != nil
}

// Clone returns a shallow copy of the block with cloned delta.
func (block *Block) Clone() *Block {
	return &Block{
		Delta:      block.Delta.Clone(),
		Heads:      block.Heads,
		Links:      block.Links,
		Encryption: block.Encryption,
	}
}

// AllLinks returns the block `Heads` and `Links` combined into a single set.
//
// All heads will be first in the set, followed by other links.
func (block *Block) AllLinks() []cidlink.Link {
	result := make([]cidlink.Link, 0, len(block.Heads)+len(block.Links))

	result = append(result, block.Heads...)

	for _, link := range block.Links {
		result = append(result, link.Link)
	}

	return result
}

// IPLDSchemaBytes returns the IPLD schema representation for the block.
//
// This needs to match the [Block] struct or [mustSetSchema] will panic on init.
func (block *Block) IPLDSchemaBytes() []byte {
	return []byte(`
		type Block struct {
			delta       CRDT
			heads       optional [Link]
			links       optional [DAGLink]
			encryption  optional Link
		}
	`)
}

// IPLDSchemaBytes returns the IPLD schema representation for the encryption block.
//
// This needs to match the [Encryption] struct or [mustSetSchema] will panic on init.
func (enc *Encryption) IPLDSchemaBytes() []byte {
	return []byte(`
		type Encryption struct {
			docID     Bytes
			fieldName optional String
			key       Bytes
		}
	`)
}

// New creates a new block with the given delta and links.
func New(delta core.Delta, links []DAGLink, heads ...cid.Cid) *Block {
	// Sort the heads lexicographically by CID.
	// We need to do this to ensure that the block is deterministic.
	sort.Slice(heads, func(i, j int) bool {
		return strings.Compare(heads[i].String(), heads[j].String()) < 0
	})

	headLinks := make([]cidlink.Link, 0, len(heads))
	for _, head := range heads {
		headLinks = append(
			headLinks,
			cidlink.Link{Cid: head},
		)
	}

	// Sort the links lexicographically by CID.
	// We need to do this to ensure that the block is deterministic.
	sort.Slice(links, func(i, j int) bool {
		return strings.Compare(links[i].Cid.String(), links[j].Cid.String()) < 0
	})

	blockLinks := make([]DAGLink, 0, len(links))
	blockLinks = append(blockLinks, links...)

	if len(headLinks) == 0 {
		// The encoding used for block serialization will consume space if an empty set is
		// provided, but it will not consume space if nil is provided, so if empty we set it
		// to nil.  The would-be space consumed includes the property name, so this is not an
		// insignificant amount.
		headLinks = nil
	}

	if len(blockLinks) == 0 {
		// The encoding used for block serialization will consume space if an empty set is
		// provided, but it will not consume space if nil is provided, so if empty we set it
		// to nil.  The would-be space consumed includes the property name, so this is not an
		// insignificant amount.
		blockLinks = nil
	}

	return &Block{
		Heads: headLinks,
		Links: blockLinks,
		Delta: crdt.NewCRDT(delta),
	}
}

// GetFromBytes returns a block from encoded bytes.
func GetEncryptionBlockFromBytes(b []byte) (*Encryption, error) {
	enc := &Encryption{}
	err := enc.Unmarshal(b)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

// GetFromBytes returns a block from encoded bytes.
func GetFromBytes(b []byte) (*Block, error) {
	block := &Block{}
	err := block.Unmarshal(b)
	if err != nil {
		return nil, err
	}
	return block, nil
}

// GetFromNode returns a block from a node.
func GetFromNode(node ipld.Node) (*Block, error) {
	block, ok := bindnode.Unwrap(node).(*Block)
	if !ok {
		return nil, NewErrNodeToBlock(node)
	}
	return block, nil
}

// GetFromNode returns a block from a node.
func GetEncryptionBlockFromNode(node ipld.Node) (*Encryption, error) {
	encBlock, ok := bindnode.Unwrap(node).(*Encryption)
	if !ok {
		return nil, NewErrNodeToBlock(node)
	}
	return encBlock, nil
}

// Marshal encodes the delta using CBOR encoding.
func (block *Block) Marshal() ([]byte, error) {
	b, err := ipld.Marshal(dagcbor.Encode, block, Schema)
	if err != nil {
		return nil, NewErrEncodingBlock(err)
	}
	return b, nil
}

// Unmarshal decodes the delta from CBOR encoding.
func (block *Block) Unmarshal(b []byte) error {
	_, err := ipld.Unmarshal(b, dagcbor.Decode, block, Schema)
	if err != nil {
		return NewErrUnmarshallingBlock(err)
	}
	return nil
}

// Marshal encodes the delta using CBOR encoding.
func (enc *Encryption) Marshal() ([]byte, error) {
	b, err := ipld.Marshal(dagcbor.Encode, enc, EncryptionSchema)
	if err != nil {
		return nil, NewErrEncodingBlock(err)
	}
	return b, nil
}

// Unmarshal decodes the delta from CBOR encoding.
func (enc *Encryption) Unmarshal(b []byte) error {
	_, err := ipld.Unmarshal(b, dagcbor.Decode, enc, EncryptionSchema)
	if err != nil {
		return NewErrUnmarshallingBlock(err)
	}
	return nil
}

// GenerateNode generates an IPLD node from the block in its representation form.
func (block *Block) GenerateNode() ipld.Node {
	return bindnode.Wrap(block, Schema).Representation()
}

// GenerateNode generates an IPLD node from the encryption block in its representation form.
func (enc *Encryption) GenerateNode() ipld.Node {
	return bindnode.Wrap(enc, EncryptionSchema).Representation()
}

// GetLinkByName returns the link by name. It will return false if the link does not exist.
func (block *Block) GetLinkByName(name string) (cidlink.Link, bool) {
	for _, link := range block.Links {
		if link.Name == name {
			return link.Link, true
		}
	}
	return cidlink.Link{}, false
}

// GenerateLink generates a cid link for the block.
func (block *Block) GenerateLink() (cidlink.Link, error) {
	node := bindnode.Wrap(block, Schema)
	return GetLinkFromNode(node.Representation())
}

// GetLinkFromNode returns the cid link from the node.
func GetLinkFromNode(node ipld.Node) (cidlink.Link, error) {
	if typedNode, ok := node.(schema.TypedNode); ok {
		node = typedNode.Representation()
	}
	lsys := cidlink.DefaultLinkSystem()
	link, err := lsys.ComputeLink(GetLinkPrototype(), node)
	if err != nil {
		return cidlink.Link{}, NewErrGeneratingLink(err)
	}
	return link.(cidlink.Link), nil
}

// GetLinkPrototype returns the link prototype for the block.
func GetLinkPrototype() cidlink.LinkPrototype {
	return cidlink.LinkPrototype{Prefix: cid.Prefix{
		Version:  uint64(multicodec.Cidv1),
		Codec:    uint64(multicodec.DagCbor),
		MhType:   uint64(multicodec.Sha2_256),
		MhLength: 32,
	}}
}
