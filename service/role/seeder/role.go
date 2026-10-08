package seeder

import (
	"context"
	"fmt"

	db "github.com/MamangRust/microservice-ecommerce-grpc-role/database/schema"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"
)

// SQL source of truth (formerly service role database/query/roles.sql).
const (
	seederGetRoles = `SELECT
    role_id,
    role_name,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER() AS total_count
FROM
    roles
WHERE
    $1::TEXT IS NULL OR role_name ILIKE '%' || $1 || '%'
ORDER BY
    created_at ASC
LIMIT $2 OFFSET $3`

	seederCreateRole = `INSERT INTO roles (
    role_name,
    created_at,
    updated_at
) VALUES (
    $1,
    current_timestamp,
    current_timestamp
) RETURNING
    role_id,
    role_name,
    created_at,
    updated_at,
    deleted_at`
)

type roleSeeder struct {
	db     *sqlx.DB
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewRoleSeeder(db *sqlx.DB, ctx context.Context, logger logger.LoggerInterface) *roleSeeder {
	return &roleSeeder{
		db:     db,
		ctx:    ctx,
		logger: logger,
	}
}

func (r *roleSeeder) Seed() error {
	// Idempotency: skip only when one of our seeded roles already exists.
	// (The e2e reset pre-inserts ROLE_ADMIN/ROLE_USER, so "any role exists"
	// must not suppress the Cashier/Manager/Admin/Supplier seeding.)
	var existing []db.GetRolesRow
	err := r.db.SelectContext(r.ctx, &existing, seederGetRoles,
		"Cashier",
		int32(1),
		int32(0),
	)
	if err == nil && len(existing) > 0 {
		r.logger.Debug("roles already seeded, skipping")
		return nil
	}

	randomRoles := []string{"Cashier", "Manager", "Admin", "Supplier"}

	totalRoles := len(randomRoles)

	for i, roleName := range randomRoles {
		_, err := r.db.ExecContext(r.ctx, seederCreateRole, roleName)
		if err != nil {
			r.logger.Error("failed to seed role", zap.Int("role", i+1), zap.String("roleName", roleName), zap.Error(err))
			return fmt.Errorf("failed to seed role %d (%s): %w", i+1, roleName, err)
		}
	}

	r.logger.Debug("role seeded successfully", zap.Int("totalRoles", totalRoles))
	return nil
}
