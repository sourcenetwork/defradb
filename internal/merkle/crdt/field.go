// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package merklecrdt

import "github.com/sourcenetwork/defradb/client"

// DocField is a struct that holds the document ID and the field value.
// This is used to have a link between the document and the field value.
// For example, to check if the field value needs to be encrypted depending on the document-level
// encryption is enabled or not.
type DocField struct {
	DocID      string
	FieldValue *client.FieldValue
}
