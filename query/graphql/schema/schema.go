// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package graphql provides the necessary schema tooling, including parsing, validation, and
// generation for developer defined types for the GraphQL implementation of DefraDB.join
package schema

import (
	"github.com/sourcenetwork/defradb/logging"
)

var (
	log = logging.MustNewLogger("defra.query.schema")
)
