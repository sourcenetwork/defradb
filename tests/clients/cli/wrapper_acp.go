// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"context"
	"encoding/json"

	"github.com/sourcenetwork/defradb/client"
)

func (w *Wrapper) AddDACPolicy(
	ctx context.Context,
	policy string,
) (client.AddPolicyResult, error) {
	args := []string{"client", "acp", "document", "policy", "add"}
	args = append(args, policy)

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return client.AddPolicyResult{}, err
	}

	var addPolicyResult client.AddPolicyResult
	if err := json.Unmarshal(data, &addPolicyResult); err != nil {
		return client.AddPolicyResult{}, err
	}

	return addPolicyResult, err
}

func (w *Wrapper) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	args := []string{
		"client", "acp", "document", "relationship", "add",
		"--collection", collectionName,
		"--docID", docID,
		"--relation", relation,
		"--actor", targetActor,
	}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return client.AddActorRelationshipResult{}, err
	}

	var exists client.AddActorRelationshipResult
	if err := json.Unmarshal(data, &exists); err != nil {
		return client.AddActorRelationshipResult{}, err
	}

	return exists, err
}

func (w *Wrapper) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	args := []string{
		"client", "acp", "document", "relationship", "delete",
		"--collection", collectionName,
		"--docID", docID,
		"--relation", relation,
		"--actor", targetActor,
	}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}

	var exists client.DeleteActorRelationshipResult
	if err := json.Unmarshal(data, &exists); err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}

	return exists, err
}

func (w *Wrapper) GetNACStatus(ctx context.Context) (client.StatusNACResult, error) {
	args := []string{"client", "acp", "node", "status"}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return client.StatusNACResult{}, err
	}

	var status client.StatusNACResult
	if err := json.Unmarshal(data, &status); err != nil {
		return client.StatusNACResult{}, err
	}

	return status, nil
}

func (w *Wrapper) ReEnableNAC(ctx context.Context) error {
	args := []string{"client", "acp", "node", "re-enable"}
	if _, err := w.cmd.execute(ctx, args); err != nil {
		return err
	}
	return nil
}

func (w *Wrapper) DisableNAC(ctx context.Context) error {
	args := []string{"client", "acp", "node", "disable"}
	if _, err := w.cmd.execute(ctx, args); err != nil {
		return err
	}
	return nil
}

func (w *Wrapper) AddNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	args := []string{
		"client", "acp", "node", "relationship", "add",
		"--relation", relation,
		"--actor", targetActor,
	}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return client.AddActorRelationshipResult{}, err
	}

	var exists client.AddActorRelationshipResult
	if err := json.Unmarshal(data, &exists); err != nil {
		return client.AddActorRelationshipResult{}, err
	}

	return exists, err
}

func (w *Wrapper) DeleteNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	args := []string{
		"client", "acp", "node", "relationship", "delete",
		"--relation", relation,
		"--actor", targetActor,
	}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}

	var exists client.DeleteActorRelationshipResult
	if err := json.Unmarshal(data, &exists); err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}

	return exists, err
}
