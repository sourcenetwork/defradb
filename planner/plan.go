// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package planner

// RequestPlan is an external hook into the planNode
// system. It allows outside packages to
// execute and manage a request plan graph directly.
// Instead of using one of the available functions
// like ExecRequest(...).
// Currently, this is used by the collection.Update
// system.
type RequestPlan planNode
