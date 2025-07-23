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
	args := []string{"client", "acp", "dac", "policy", "add"}
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
		"client", "acp", "dac", "relationship", "add",
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
		"client", "acp", "dac", "relationship", "delete",
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

func (w *Wrapper) GetAACStatus(ctx context.Context) (client.StatusAACResult, error) {
	args := []string{"client", "acp", "aac", "status"}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return client.StatusAACResult{}, err
	}

	var status client.StatusAACResult
	if err := json.Unmarshal(data, &status); err != nil {
		return client.StatusAACResult{}, err
	}

	return status, nil
}

func (w *Wrapper) ReEnableAAC(ctx context.Context) error {
	args := []string{"client", "acp", "aac", "re-enable"}
	if _, err := w.cmd.execute(ctx, args); err != nil {
		return err
	}
	return nil
}

func (w *Wrapper) DisableAAC(ctx context.Context) error {
	args := []string{"client", "acp", "aac", "disable"}
	if _, err := w.cmd.execute(ctx, args); err != nil {
		return err
	}
	return nil
}

func (w *Wrapper) AddAACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	args := []string{
		"client", "acp", "aac", "relationship", "add",
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

func (w *Wrapper) DeleteAACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	args := []string{
		"client", "acp", "aac", "relationship", "delete",
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
