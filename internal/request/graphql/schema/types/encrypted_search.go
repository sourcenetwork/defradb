// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package types

import (
	"github.com/sourcenetwork/defradb/client/request"
	gql "github.com/sourcenetwork/graphql-go"
)

const (
	encryptedSearchResultDescription = "Result type for encrypted search queries containing matching document IDs"
	docIDsFieldDescription           = "Array of document IDs matching the encrypted search criteria"
)

// EncryptedSearchResultObject creates the GraphQL object type for encrypted search results.
// This is a shared type used by all collections with encrypted indexes.
//
//	type EncryptedSearchResult {
//		docIDs: [String!]!
//	}
func EncryptedSearchResultObject() *gql.Object {
	return gql.NewObject(gql.ObjectConfig{
		Name:        request.EncryptedSearchResultName,
		Description: encryptedSearchResultDescription,
		Fields: gql.Fields{
			request.DocIDsFieldName: &gql.Field{
				Description: docIDsFieldDescription,
				Type:        gql.NewNonNull(gql.NewList(gql.NewNonNull(gql.String))),
			},
		},
	})
}
