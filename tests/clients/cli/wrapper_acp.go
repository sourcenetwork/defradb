// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package cli

import (
	"context"
	"encoding/json"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/utils"
)

func (w *Wrapper) AddDACPolicy(
	ctx context.Context,
	policy string,
	opts ...options.Enumerable[options.AddDACPolicyOptions],
) (client.AddPolicyResult, error) {
	args := []string{"client", "acp", "document", "policy", "add"}
	args = append(args, policy)

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

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
	opts ...options.Enumerable[options.AddDACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	args := []string{
		"client", "acp", "document", "relationship", "add",
		"--collection", collectionName,
		"--docID", docID,
		"--relation", relation,
		"--actor", targetActor,
	}

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

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
	opts ...options.Enumerable[options.DeleteDACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	args := []string{
		"client", "acp", "document", "relationship", "delete",
		"--collection", collectionName,
		"--docID", docID,
		"--relation", relation,
		"--actor", targetActor,
	}

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

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

func (w *Wrapper) GetNACStatus(
	ctx context.Context,
	opts ...options.Enumerable[options.GetNACStatusOptions],
) (client.NACStatusResult, error) {
	args := []string{"client", "acp", "node", "status"}

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return client.NACStatusResult{}, err
	}

	var status client.NACStatusResult
	if err := json.Unmarshal(data, &status); err != nil {
		return client.NACStatusResult{}, err
	}

	return status, nil
}

func (w *Wrapper) ReEnableNAC(
	ctx context.Context,
	opts ...options.Enumerable[options.ReEnableNACOptions],
) error {
	args := []string{"client", "acp", "node", "re-enable"}

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

	if _, err := w.cmd.execute(ctx, args); err != nil {
		return err
	}
	return nil
}

func (w *Wrapper) DisableNAC(
	ctx context.Context,
	opts ...options.Enumerable[options.DisableNACOptions],
) error {
	args := []string{"client", "acp", "node", "disable"}

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

	if _, err := w.cmd.execute(ctx, args); err != nil {
		return err
	}
	return nil
}

func (w *Wrapper) AddNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.AddNACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	args := []string{
		"client", "acp", "node", "relationship", "add",
		"--relation", relation,
		"--actor", targetActor,
	}

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

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
	opts ...options.Enumerable[options.DeleteNACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	args := []string{
		"client", "acp", "node", "relationship", "delete",
		"--relation", relation,
		"--actor", targetActor,
	}

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

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
