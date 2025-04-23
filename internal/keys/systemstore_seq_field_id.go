// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keys

import (
	ds "github.com/ipfs/go-datastore"
)

// FieldIDSequenceKey is used to key the sequence used to generate field ids.
//
// The sequence is specific to each collection root.  Multiple collection of the same root
// must maintain consistent field ids.
type FieldIDSequenceKey struct {
	SchemaRoot string
}

var _ Key = (*FieldIDSequenceKey)(nil)

func NewFieldIDSequenceKey(schemaRoot string) FieldIDSequenceKey {
	return FieldIDSequenceKey{SchemaRoot: schemaRoot}
}

func (k FieldIDSequenceKey) ToString() string {
	return FIELD_ID_SEQ + "/" + k.SchemaRoot
}

func (k FieldIDSequenceKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k FieldIDSequenceKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
