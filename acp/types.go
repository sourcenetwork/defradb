// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package acp

// RegistrationResult is an enum type which indicates the result of a RegisterObject call to SourceHub / ACP Core
type RegistrationResult int32

const (
	// NoOp indicates no action was take. The operation failed or the Object already existed and was active
	RegistrationResult_NoOp RegistrationResult = 0
	// Registered indicates the Object was sucessfuly registered to the Actor.
	RegistrationResult_Registered RegistrationResult = 1
	// Unarchived indicates that a previously deleted Object is active again.
	// Only the original owners can Unarchive an object.
	RegistrationResult_Unarchived RegistrationResult = 2
)

// policy is a data container carrying the necessary data
// to verify whether a policy meets DPI requirements
type policy struct {
	ID        string
	Resources map[string]*resource
}

// resource is a data container carrying the necessary data
// to verify whether it meets DPI requirements.
type resource struct {
	Name        string
	Permissions map[string]*permission
}

// permission is a data container carrying the necessary data
// to verify whether it meets DPI requirements.
type permission struct {
	Name       string
	Expression string
}
