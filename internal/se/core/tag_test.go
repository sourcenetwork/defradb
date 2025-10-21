// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package secore

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEqualityTag_WhenGenerated_Returns16ByteTag(t *testing.T) {
	key := []byte("test-key-32-bytes-long-for-hmac!")
	collectionID := "collection123"
	fieldName := "email"
	value := []byte("user@example.com")

	tag := GenerateEqualityTag(key, "", collectionID, fieldName, value)
	require.Len(t, tag, 16, "tag should be 16 bytes (truncated from 32)")
}

func TestEqualityTag_WithSameInputs_ReturnsSameTag(t *testing.T) {
	key := []byte("test-key-32-bytes-long-for-hmac!")
	collectionID := "collection123"
	fieldName := "email"
	value := []byte("user@example.com")

	tag1 := GenerateEqualityTag(key, "", collectionID, fieldName, value)

	tag2 := GenerateEqualityTag(key, "", collectionID, fieldName, value)

	require.Equal(t, tag1, tag2, "same inputs should produce same tag")
}

func TestEqualityTag_WhenOnlyFieldNameDiffers_ReturnsDifferentTag(t *testing.T) {
	key := []byte("test-key-32-bytes-long-for-hmac!")
	collectionID := "collection123"
	value := []byte("user@example.com")

	tagEmail := GenerateEqualityTag(key, "", collectionID, "email", value)

	tagName := GenerateEqualityTag(key, "", collectionID, "name", value)

	require.NotEqual(t, tagEmail, tagName, "tags should be different for different fields")
}

func TestEqualityTag_WhenOnlyCollectionDiffers_ReturnsDifferentTag(t *testing.T) {
	key := []byte("test-key-32-bytes-long-for-hmac!")
	fieldName := "email"
	value := []byte("user@example.com")

	tagColl1 := GenerateEqualityTag(key, "", "collection123", fieldName, value)

	tagColl2 := GenerateEqualityTag(key, "", "collection456", fieldName, value)

	require.NotEqual(t, tagColl1, tagColl2, "tags should be different for different collections")
}

func TestEqualityTag_WhenOnlyValueDiffers_ReturnsDifferentTag(t *testing.T) {
	key := []byte("test-key-32-bytes-long-for-hmac!")
	collectionID := "collection123"
	fieldName := "email"

	tag1 := GenerateEqualityTag(key, "", collectionID, fieldName, []byte("user@example.com"))

	tag2 := GenerateEqualityTag(key, "", collectionID, fieldName, []byte("different@example.com"))

	require.NotEqual(t, tag1, tag2, "tags should be different for different values")
}

func TestEqualityTag_WhenOnlyKeyDiffers_ReturnsDifferentTag(t *testing.T) {
	collectionID := "collection123"
	fieldName := "email"
	value := []byte("user@example.com")

	tag1 := GenerateEqualityTag([]byte("key1-32-bytes-long-for-hmac-123"), "", collectionID, fieldName, value)

	tag2 := GenerateEqualityTag([]byte("key2-32-bytes-long-for-hmac-456"), "", collectionID, fieldName, value)

	require.NotEqual(t, tag1, tag2, "tags should be different for different keys")
}

func TestEqualityTag_WithEmptyValue_ReturnsValidTag(t *testing.T) {
	key := []byte("test-key-32-bytes-long-for-hmac!")
	collectionID := "collection123"
	fieldName := "email"

	tag := GenerateEqualityTag(key, "", collectionID, fieldName, []byte(""))
	require.NotNil(t, tag)
	require.Len(t, tag, 16)
}

func TestEqualityTag_WithNilValue_ReturnsValidTag(t *testing.T) {
	key := []byte("test-key-32-bytes-long-for-hmac!")
	collectionID := "collection123"
	fieldName := "email"

	tag := GenerateEqualityTag(key, "", collectionID, fieldName, nil)
	require.NotNil(t, tag)
	require.Len(t, tag, 16)
}
