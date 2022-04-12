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
Package logging abstracts away any underlying logging library providing
a single contact-point for the dependency allowing relatively easy
swapping out should we want to.

This package allows configuration to be loaded and globally applied
after logger instances have been created, utilising an internal thread-safe
registry of named logger instances to apply the config to.

Configuration may be applied globally, or to logger instances of a specific
name, with the named-configuration being used over the global settings if
both are provided.

All configuration options are optional.
*/
package logging
