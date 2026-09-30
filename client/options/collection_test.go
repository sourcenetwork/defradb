// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package options

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/internal/utils"
)

func TestDocumentOptionsSetEnableSigning(t *testing.T) {
	add := utils.NewOptions(AddDocument().SetEnableSigning(false))
	require.True(t, add.EnableSigning.HasValue())
	require.False(t, add.EnableSigning.Value())

	update := utils.NewOptions(UpdateDocument().SetEnableSigning(true))
	require.True(t, update.EnableSigning.HasValue())
	require.True(t, update.EnableSigning.Value())

	deleteOpts := utils.NewOptions(DeleteDocument().SetEnableSigning(false))
	require.True(t, deleteOpts.EnableSigning.HasValue())
	require.False(t, deleteOpts.EnableSigning.Value())

	updateWithFilter := utils.NewOptions(UpdateDocumentsWithFilter().SetEnableSigning(true))
	require.True(t, updateWithFilter.EnableSigning.HasValue())
	require.True(t, updateWithFilter.EnableSigning.Value())

	deleteWithFilter := utils.NewOptions(DeleteDocumentsWithFilter().SetEnableSigning(false))
	require.True(t, deleteWithFilter.EnableSigning.HasValue())
	require.False(t, deleteWithFilter.EnableSigning.Value())
}

func TestAddDocumentOptionsCopiesEncryptedFields(t *testing.T) {
	fields := []string{"name", "age"}
	add := utils.NewOptions(AddDocument().SetEncryptedFields(fields))

	fields[0] = "mutated"

	require.Equal(t, []string{"name", "age"}, add.EncryptedFields)
}
