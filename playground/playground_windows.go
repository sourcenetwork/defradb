// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build playground

//go:generate powershell -ExecutionPolicy Bypass -File ../tools/scripts/download_playground.ps1

package playground

import (
	"embed"
)

//go:embed dist
var Dist embed.FS
