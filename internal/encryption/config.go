// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package encryption

// DocEncConfig is the configuration for document encryption.
type DocEncConfig struct {
	// IsEncrypted is a flag to indicate if the document should be encrypted.
	IsEncrypted bool
	//  EncryptedFields is a list of fields individual that should be encrypted.
	EncryptedFields []string
}
