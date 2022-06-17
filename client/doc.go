// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package client provides public members for interacting with a Defra DB instance.

Only calls made via the `DB` and `Collection` interfaces interact with the underlying datastores. Currently the only
provided implementation of `DB` is found in the `defra/db` package and can be instantiated via the `NewDB` function.
*/
package client
