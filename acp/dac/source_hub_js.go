// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// SourceHub ACP implementation for JS/WASM environments.
//
// Client applications must include the acp-js library, which exposes the following
// bridge functions that interface with SourceHub ACP:
//
//   - acp_AddPolicy(policy: string, policyMarshalType: number) -> Promise<[string | null, Error | null]>
//     Adds a new access control policy and returns its ID
//
//   - acp_Policy(policyId: string) -> Promise<[string | null, Error | null]>
//     Retrieves a policy by ID, returning its JSON string representation
//
//   - acp_RegisterObject(policyId: string, resourceName: string, objectId: string) -> Promise<[any | null, Error | null]>
//     Registers an object (document) with a policy resource for access control
//
//   - acp_ObjectOwner(policyId: string, resourceName: string, objectId: string) -> Promise<[string | null, Error | null]>
//     Returns the owner identity of a registered object
//
//   - acp_VerifyAccessRequest(permission: string, actorId: string, policyId: string, resourceName: string, objectId: string) -> Promise<[boolean, Error | null]>
//     Verifies if an actor has the specified permission on an object
//
//   - acp_AddActorRelationship(policyId: string, resourceName: string, objectId: string, relation: string, targetActor: string) -> Promise<[boolean | null, Error | null]>
//     Adds a relationship between an actor and an object
//
//   - acp_DeleteActorRelationship(policyId: string, resourceName: string, objectId: string, relation: string, targetActor: string) -> Promise<[boolean | null, Error | null]>
//     Removes a relationship between an actor and an object
//
//go:build js

package dac

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	sysjs "syscall/js"

	protoTypes "github.com/cosmos/gogoproto/types"
	"github.com/sourcenetwork/goji"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
)

type SourceHubDocumentACP struct{}

// NewSourceHubDocumentACP returns a new SourceHub ACP instance for JavaScript environments.
func NewSourceHubDocumentACP() DocumentACP {
	return &bridgeDocumentACP{
		clientACP: &SourceHubDocumentACP{},
	}
}

// callJSFunction calls JavaScript functions and handles async results
func callJSFunction(funcName string, args ...interface{}) ([]sysjs.Value, error) {
	funcObj := sysjs.Global().Get(funcName)
	if !funcObj.Truthy() {
		return nil, fmt.Errorf(
			"JavaScript function %s not found",
			funcName,
		)
	}
	promise := funcObj.Invoke(args...)
	if !promise.Truthy() {
		return nil, fmt.Errorf(
			"JavaScript function %s returned null or undefined",
			funcName,
		)
	}
	results, err := goji.Await(goji.PromiseValue(promise))
	if err != nil {
		return nil, fmt.Errorf("failed to await %s: %w", funcName, err)
	}
	return results, nil
}

// fromSourceHubPolicyJS converts a JavaScript policy object to acpTypes.Policy
func fromSourceHubPolicyJS(jsPolicy map[string]interface{}) acpTypes.Policy {
	resources := make(map[string]*acpTypes.Resource)
	if resourcesData, ok := jsPolicy["resources"].([]interface{}); ok {
		for _, resourceData := range resourcesData {
			if resourceMap, ok := resourceData.(map[string]interface{}); ok {
				if resourceName, ok := resourceMap["name"].(string); ok {
					resource := fromSourceHubResourceJS(resourceName, resourceMap)
					resources[resource.Name] = resource
				}
			}
		}
	} else if resourcesData, ok := jsPolicy["resources"].(map[string]interface{}); ok {
		for resourceName, resourceData := range resourcesData {
			if resourceMap, ok := resourceData.(map[string]interface{}); ok {
				resource := fromSourceHubResourceJS(resourceName, resourceMap)
				resources[resource.Name] = resource
			}
		}
	}
	return acpTypes.Policy{
		ID:        jsPolicy["id"].(string),
		Resources: resources,
	}
}

// fromSourceHubResourceJS converts a JavaScript resource object to acpTypes.Resource
func fromSourceHubResourceJS(resourceName string, resourceMap map[string]interface{}) *acpTypes.Resource {
	perms := make(map[string]*acpTypes.Permission)
	if permissionsData, ok := resourceMap["permissions"].([]interface{}); ok {
		for _, permData := range permissionsData {
			if permMap, ok := permData.(map[string]interface{}); ok {
				if permName, ok := permMap["name"].(string); ok {
					perm := fromSourceHubPermissionJS(permName, permMap)
					perms[perm.Name] = perm
				}
			}
		}
	} else if permissionsData, ok := resourceMap["permissions"].(map[string]interface{}); ok {
		for permName, permData := range permissionsData {
			if permMap, ok := permData.(map[string]interface{}); ok {
				perm := fromSourceHubPermissionJS(permName, permMap)
				perms[perm.Name] = perm
			}
		}
	}
	return &acpTypes.Resource{
		Name:        resourceName,
		Permissions: perms,
	}
}

// fromSourceHubPermissionJS converts a JavaScript permission object to acpTypes.Permission
func fromSourceHubPermissionJS(permName string, permMap map[string]interface{}) *acpTypes.Permission {
	expression := ""
	if expr, ok := permMap["expr"].(string); ok {
		expression = expr
	} else if expr, ok := permMap["expression"].(string); ok {
		expression = expr
	}
	expression = strings.TrimSpace(expression)
	if strings.HasPrefix(expression, "(") && strings.HasSuffix(expression, ")") {
		parenCount := 0
		balanced := true
		for i, char := range expression {
			if char == '(' {
				parenCount++
			} else if char == ')' {
				parenCount--
				if parenCount == 0 && i < len(expression)-1 {
					balanced = false
					break
				}
			}
		}
		if balanced && parenCount == 0 {
			expression = expression[1 : len(expression)-1]
		}
	}
	return &acpTypes.Permission{
		Name:       permName,
		Expression: expression,
	}
}

func (a *SourceHubDocumentACP) Start(ctx context.Context) error {
	// No-op: client is initialized during node creation
	return nil
}

func (a *SourceHubDocumentACP) AddPolicy(
	ctx context.Context,
	creator identity.Identity,
	policy string,
	policyMarshalType acpTypes.PolicyMarshalType,
	creationTime *protoTypes.Timestamp,
) (string, error) {
	results, err := callJSFunction("acp_AddPolicy", policy, int(policyMarshalType))
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return "", fmt.Errorf("AddPolicy returned no results")
	}
	var policyID string
	if results[0].Type() == sysjs.TypeObject {
		firstElement := results[0].Index(0)
		if firstElement.Truthy() {
			policyID = firstElement.String()
		} else {
			return "", fmt.Errorf(
				"AddPolicy returned object but first element is not truthy",
			)
		}
	} else {
		policyID = results[0].String()
	}
	if policyID == "" {
		return "", fmt.Errorf("AddPolicy returned empty policy ID")
	}
	return policyID, nil
}

func (a *SourceHubDocumentACP) Policy(
	ctx context.Context,
	policyID string,
) (immutable.Option[acpTypes.Policy], error) {
	results, err := callJSFunction("acp_Policy", policyID)
	if err != nil {
		return immutable.None[acpTypes.Policy](), err
	}
	if len(results) == 0 || !results[0].Truthy() {
		return immutable.None[acpTypes.Policy](), nil
	}
	var policyObj sysjs.Value
	if results[0].Type() == sysjs.TypeObject && results[0].InstanceOf(sysjs.Global().Get("Array")) {
		if results[0].Length() > 0 {
			policyObj = results[0].Index(0)
		} else {
			return immutable.None[acpTypes.Policy](), nil
		}
	} else {
		policyObj = results[0]
	}
	if !policyObj.Truthy() {
		return immutable.None[acpTypes.Policy](), nil
	}
	var policyStr string
	if policyObj.Type() == sysjs.TypeString {
		policyStr = policyObj.String()
	} else {
		jsonStr := sysjs.Global().Get("JSON").Call("stringify", policyObj)
		if !jsonStr.Truthy() {
			return immutable.None[acpTypes.Policy](), nil
		}
		policyStr = jsonStr.String()
	}
	if policyStr == "" || policyStr == "null" {
		return immutable.None[acpTypes.Policy](), nil
	}
	var jsPolicy map[string]interface{}
	if err := json.Unmarshal([]byte(policyStr), &jsPolicy); err != nil {
		return immutable.None[acpTypes.Policy](), err
	}
	policy := fromSourceHubPolicyJS(jsPolicy)
	return immutable.Some(policy), nil
}

func (a *SourceHubDocumentACP) RegisterObject(
	ctx context.Context,
	id identity.Identity,
	policyID string,
	resourceName string,
	objectID string,
	creationTime *protoTypes.Timestamp,
) error {
	_, err := callJSFunction("acp_RegisterObject", policyID, resourceName, objectID)
	return err
}

func (a *SourceHubDocumentACP) ObjectOwner(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
) (immutable.Option[string], error) {
	results, err := callJSFunction("acp_ObjectOwner", policyID, resourceName, objectID)
	if err != nil {
		return immutable.None[string](), err
	}
	if len(results) > 0 && results[0].Truthy() && results[0].String() != "" {
		return immutable.Some(results[0].String()), nil
	}
	return immutable.None[string](), nil
}

func (a *SourceHubDocumentACP) VerifyAccessRequest(
	ctx context.Context,
	permission acpTypes.ResourceInterfacePermission,
	actorID string,
	policyID string,
	resourceName string,
	objectID string,
) (bool, error) {
	results, err := callJSFunction(
		"acp_VerifyAccessRequest",
		permission.String(),
		actorID,
		policyID,
		resourceName,
		objectID,
	)
	if err != nil {
		return false, err
	}
	if len(results) == 0 {
		return false, nil
	}
	result := results[0]
	if !result.Truthy() {
		return false, nil
	}
	switch result.Type() {
	case sysjs.TypeBoolean:
		return result.Bool(), nil
	case sysjs.TypeObject:
		if result.InstanceOf(sysjs.Global().Get("Array")) && result.Length() > 0 {
			firstElement := result.Index(0)
			if firstElement.Type() == sysjs.TypeBoolean {
				return firstElement.Bool(), nil
			}
		}
		if resultProp := result.Get("result"); !resultProp.IsUndefined() && resultProp.Type() == sysjs.TypeBoolean {
			return resultProp.Bool(), nil
		}
		if successProp := result.Get("success"); !successProp.IsUndefined() && successProp.Type() == sysjs.TypeBoolean {
			return successProp.Bool(), nil
		}
		return false, nil
	default:
		return false, nil
	}
}

func (a *SourceHubDocumentACP) Close() error {
	return nil
}

func (a *SourceHubDocumentACP) ResetState(_ context.Context) error {
	return nil
}

func (a *SourceHubDocumentACP) AddActorRelationship(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
	relation string,
	requester identity.Identity,
	targetActor string,
	creationTime *protoTypes.Timestamp,
) (bool, error) {
	results, err := callJSFunction(
		"acp_AddActorRelationship",
		policyID,
		resourceName,
		objectID,
		relation,
		targetActor,
	)
	if err != nil {
		return false, err
	}
	if len(results) == 0 {
		return false, fmt.Errorf("AddActorRelationship returned no results")
	}
	result := results[0]
	if !result.Truthy() {
		return false, nil
	}
	if result.Type() == sysjs.TypeObject && result.InstanceOf(sysjs.Global().Get("Array")) {
		if result.Length() >= 2 {
			errorVal := result.Index(1)
			if errorVal.Truthy() && errorVal.Type() == sysjs.TypeObject {
				errorStr := ""
				if errorVal.Type() == sysjs.TypeString {
					errorStr = errorVal.String()
				} else {
					if messageProp := errorVal.Get("message"); messageProp.Truthy() {
						errorStr = messageProp.String()
					} else {
						errorStr = "Unknown error"
					}
				}
				return false, fmt.Errorf("AddActorRelationship error: %s", errorStr)
			}
			boolVal := result.Index(0)
			if boolVal.Type() == sysjs.TypeBoolean {
				boolResult := boolVal.Bool()
				return boolResult, nil
			} else if boolVal.Type() == sysjs.TypeNull || !boolVal.Truthy() {
				return false, nil
			}
		}
	}
	return false, nil
}

func (a *SourceHubDocumentACP) DeleteActorRelationship(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
	relation string,
	requester identity.Identity,
	targetActor string,
	creationTime *protoTypes.Timestamp,
) (bool, error) {
	results, err := callJSFunction(
		"acp_DeleteActorRelationship",
		policyID,
		resourceName,
		objectID,
		relation,
		targetActor,
	)
	if err != nil {
		return false, err
	}
	if len(results) == 0 {
		return false, fmt.Errorf("DeleteActorRelationship returned no results")
	}
	result := results[0]
	if !result.Truthy() {
		return false, nil
	}
	if result.Type() == sysjs.TypeObject && result.InstanceOf(sysjs.Global().Get("Array")) {
		if result.Length() >= 2 {
			errorVal := result.Index(1)
			if errorVal.Truthy() && errorVal.Type() == sysjs.TypeObject {
				errorStr := ""
				if errorVal.Type() == sysjs.TypeString {
					errorStr = errorVal.String()
				} else {
					if messageProp := errorVal.Get("message"); messageProp.Truthy() {
						errorStr = messageProp.String()
					} else {
						errorStr = "Unknown error"
					}
				}
				return false, fmt.Errorf("DeleteActorRelationship error: %s", errorStr)
			}
			boolVal := result.Index(0)
			if boolVal.Type() == sysjs.TypeBoolean {
				boolResult := boolVal.Bool()
				return boolResult, nil
			} else if boolVal.Type() == sysjs.TypeNull || !boolVal.Truthy() {
				return false, nil
			}
		}
	}
	return false, nil
}
