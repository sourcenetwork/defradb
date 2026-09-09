// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.

package schema

import _ "embed"

// defaultSchemaSDL is the canonical SDL representation of DefraDB's fixed
// GraphQL schema.
//
//go:embed templates/default.graphql
var defaultSchemaSDL string
