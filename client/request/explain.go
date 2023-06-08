// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package request

// ExplainType does not represent which type is currently the default explain request type.
type ExplainType string

// Types of explain requests.
const (
	SimpleExplain  ExplainType = "simple"
	ExecuteExplain ExplainType = "execute"
	DebugExplain   ExplainType = "debug"
)
