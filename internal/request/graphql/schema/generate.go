// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"context"

	"github.com/sourcenetwork/defradb/client"
)

const (
	filterInputNameSuffix          = "FilterArg"
	encryptedFilterInputNameSuffix = "EncryptedFilterArg"
)

// Generate renders and validates a complete schema for collections before
// atomically replacing the manager's current definition.
func (s *SchemaManager) Generate(_ context.Context, collections []client.CollectionVersion) error {
	sdl, err := renderSchemaSDL(collections, s.isSearchableEncryptionEnabled)
	if err != nil {
		return err
	}
	if err := s.setDefinition(sdl); err != nil {
		return err
	}
	return nil
}
