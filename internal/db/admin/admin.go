// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package admin

import (
	"context"

	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/acp/aac"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db"
)

var (
	log = corelog.NewLogger("db_admin")

	_ client.DB = (*AdminDB)(nil)
)

// AdminDB is admin access controlled database.
type AdminDB struct {
	// Contains admin access information.
	adminInfo *db.AdminInfo

	// Normal database instance (without any admin access control).
	db *db.DB
}

// NewAdminDB creates a new instance of the AdminDB..
func NewAdminDB(
	ctx context.Context,
	adminInfo *db.AdminInfo,
	db *db.DB,
) (*AdminDB, error) {
	return newAdminDB(ctx, adminInfo, db)
}

func newAdminDB(
	ctx context.Context,
	adminInfo *db.AdminInfo,
	db *db.DB,
) (*AdminDB, error) {
	adminDB := &AdminDB{
		adminInfo: adminInfo,
		db:        db,
	}

	// TODO-ACP-ADMIN: FIX logic in separate admin package
	// Start admin acp if enabled, this will recover previous state if there is any.
	// to free resources must call [adminDB.Close()] when done.
	if err := adminInfo.AdminACP.Start(ctx); err != nil {
		return nil, err
	}

	return adminDB, nil
}

// checkAdminAccess is a helper function that performs the admin acp validation check and
// executes the operation specific logic, if identity has admin access for that operation.
// It uses a closure to handle arbitrary method signatures.
func checkAdminAccess(
	ctx context.Context,
	adminACP aac.AdminACP,
	permissionNeeded acpTypes.AdminResourcePermission,
) (bool, error) {
	// Extract admin identity from ctx

	return false, nil
}

//func (adminDB *AdminDB) checkAdminAccess(
//	ctx,
//	permissionNeeded acpTypes.AdminResourcePermission,
//	action func() error,
//) bool, error {
//	return checkAdminAcess(ctx, func() error, action func() error)
//}
