// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package hnsw

import (
	"encoding/binary"
	"errors"
	"math"
)

// nodeEncodingVersion and metaEncodingVersion are the current byte-layout
// versions for MarshalNode/MarshalMeta respectively. They are written as
// the first byte of the encoded output so that future layout changes can
// be detected and, if needed, migrated.
const (
	nodeEncodingVersion byte = 0x01
	metaEncodingVersion byte = 0x01
)

// Field widths (in bytes) of the fixed-size components shared by the Node and Meta layouts. Named
// so the size calculations and the offset advances cannot drift apart from the documented layout.
const (
	versionWidth   = 1 // version byte
	flagWidth      = 1 // a boolean flag (Deleted / Empty)
	countWidth     = 4 // a uint32 length/count prefix (vector len, layer count, neighbour count)
	float32Width   = 4 // a single float32 vector element
	nodeIDWidth    = 8 // a NodeID (uint64): node ID, neighbour ID, entry point
	topLayerWidth  = 4 // Meta.TopLayer (int32)
	metaEncodedLen = versionWidth + nodeIDWidth + topLayerWidth + flagWidth
	// nodeHeaderLen is the smallest valid encoded node: version + ID + Deleted + vector-length +
	// layer-count, before any vector elements or layers.
	nodeHeaderLen = versionWidth + nodeIDWidth + flagWidth + countWidth + countWidth
)

// ErrInvalidNodeEncoding is returned by UnmarshalNode when the input bytes
// are truncated, corrupt, or use an unsupported encoding version.
var ErrInvalidNodeEncoding = errors.New("hnsw: invalid node encoding")

// ErrInvalidMetaEncoding is returned by UnmarshalMeta when the input bytes
// are truncated, corrupt, or use an unsupported encoding version.
var ErrInvalidMetaEncoding = errors.New("hnsw: invalid meta encoding")

// MarshalNode encodes n into a compact, deterministic byte representation
// suitable for storage as a single KV value.
//
// Layout (little-endian):
//
//	1 byte   version (nodeEncodingVersion)
//	8 bytes  ID (uint64)
//	1 byte   Deleted (0 or 1)
//	4 bytes  vector length (uint32)
//	4 bytes  * vector length: vector elements (float32 bits)
//	4 bytes  number of layers (uint32)
//	for each layer:
//	  4 bytes  neighbour count (uint32)
//	  8 bytes  * neighbour count: neighbour ids (uint64)
func MarshalNode(n Node) ([]byte, error) {
	size := nodeHeaderLen + float32Width*len(n.Vector)
	for _, layer := range n.Layers {
		size += countWidth + nodeIDWidth*len(layer)
	}

	buf := make([]byte, size)
	off := 0

	buf[off] = nodeEncodingVersion
	off += versionWidth

	binary.LittleEndian.PutUint64(buf[off:], uint64(n.ID))
	off += nodeIDWidth

	if n.Deleted {
		buf[off] = 1
	} else {
		buf[off] = 0
	}
	off += flagWidth

	binary.LittleEndian.PutUint32(buf[off:], uint32(len(n.Vector)))
	off += countWidth
	for _, x := range n.Vector {
		binary.LittleEndian.PutUint32(buf[off:], math.Float32bits(x))
		off += float32Width
	}

	binary.LittleEndian.PutUint32(buf[off:], uint32(len(n.Layers)))
	off += countWidth
	for _, layer := range n.Layers {
		binary.LittleEndian.PutUint32(buf[off:], uint32(len(layer)))
		off += countWidth
		for _, id := range layer {
			binary.LittleEndian.PutUint64(buf[off:], uint64(id))
			off += nodeIDWidth
		}
	}

	return buf, nil
}

// UnmarshalNode decodes b, previously produced by MarshalNode, into a
// Node. It returns ErrInvalidNodeEncoding if b is truncated, corrupt, or
// uses an unsupported encoding version.
func UnmarshalNode(b []byte) (Node, error) {
	if len(b) < nodeHeaderLen {
		return Node{}, ErrInvalidNodeEncoding
	}
	off := 0

	version := b[off]
	off += versionWidth
	if version != nodeEncodingVersion {
		return Node{}, ErrInvalidNodeEncoding
	}

	id := NodeID(binary.LittleEndian.Uint64(b[off:]))
	off += nodeIDWidth

	deleted := b[off] != 0
	off += flagWidth

	vecLen := binary.LittleEndian.Uint32(b[off:])
	off += countWidth
	if uint64(off)+uint64(vecLen)*float32Width > uint64(len(b)) {
		return Node{}, ErrInvalidNodeEncoding
	}
	var vector []float32
	if vecLen > 0 {
		vector = make([]float32, vecLen)
		for i := range vector {
			vector[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[off:]))
			off += float32Width
		}
	}

	if len(b) < off+countWidth {
		return Node{}, ErrInvalidNodeEncoding
	}
	numLayers := binary.LittleEndian.Uint32(b[off:])
	off += countWidth

	var layers [][]NodeID
	if numLayers > 0 {
		layers = make([][]NodeID, numLayers)
		for i := range layers {
			if len(b) < off+countWidth {
				return Node{}, ErrInvalidNodeEncoding
			}
			neighbourCount := binary.LittleEndian.Uint32(b[off:])
			off += countWidth
			if uint64(off)+uint64(neighbourCount)*nodeIDWidth > uint64(len(b)) {
				return Node{}, ErrInvalidNodeEncoding
			}
			neighbours := make([]NodeID, neighbourCount)
			for j := range neighbours {
				neighbours[j] = NodeID(binary.LittleEndian.Uint64(b[off:]))
				off += nodeIDWidth
			}
			layers[i] = neighbours
		}
	}

	if off != len(b) {
		return Node{}, ErrInvalidNodeEncoding
	}

	return Node{
		ID:      id,
		Vector:  vector,
		Layers:  layers,
		Deleted: deleted,
	}, nil
}

// MarshalMeta encodes m into a compact, deterministic byte representation
// suitable for storage as a single KV value.
//
// Layout (little-endian):
//
//	1 byte  version (metaEncodingVersion)
//	8 bytes EntryPoint (uint64)
//	4 bytes TopLayer (int32)
//	1 byte  Empty (0 or 1)
func MarshalMeta(m Meta) ([]byte, error) {
	buf := make([]byte, metaEncodedLen)
	off := 0

	buf[off] = metaEncodingVersion
	off += versionWidth

	binary.LittleEndian.PutUint64(buf[off:], uint64(m.EntryPoint))
	off += nodeIDWidth

	binary.LittleEndian.PutUint32(buf[off:], uint32(int32(m.TopLayer)))
	off += topLayerWidth

	if m.Empty {
		buf[off] = 1
	} else {
		buf[off] = 0
	}

	return buf, nil
}

// UnmarshalMeta decodes b, previously produced by MarshalMeta, into a
// Meta. It returns ErrInvalidMetaEncoding if b is truncated, corrupt, or
// uses an unsupported encoding version.
func UnmarshalMeta(b []byte) (Meta, error) {
	if len(b) != metaEncodedLen {
		return Meta{}, ErrInvalidMetaEncoding
	}
	off := 0

	version := b[off]
	off += versionWidth
	if version != metaEncodingVersion {
		return Meta{}, ErrInvalidMetaEncoding
	}

	entryPoint := NodeID(binary.LittleEndian.Uint64(b[off:]))
	off += nodeIDWidth

	topLayer := int(int32(binary.LittleEndian.Uint32(b[off:])))
	off += topLayerWidth

	empty := b[off] != 0

	return Meta{
		EntryPoint: entryPoint,
		TopLayer:   topLayer,
		Empty:      empty,
	}, nil
}
