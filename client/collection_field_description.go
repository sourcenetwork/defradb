// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import "fmt"

// FieldID is a unique identifier for a field in a schema.
type FieldID uint32

// CollectionFieldDescription describes the local components of a field on a collection.
type CollectionFieldDescription struct {
	// Name contains the name of the [SchemaFieldDescription] that this field uses.
	Name string

	// ID contains the local, internal ID of this field.
	ID FieldID
}

func (f FieldID) String() string {
	return fmt.Sprint(uint32(f))
}
