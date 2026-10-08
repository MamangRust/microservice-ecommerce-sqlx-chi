package seeder

import (
	"context"
	"fmt"
	"time"

	db "github.com/MamangRust/microservice-ecommerce-grpc-role/database/schema"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"
	"golang.org/x/exp/rand"
)

// SQL source of truth (formerly service role database/query/roles.sql and
// database/query/users_role.sql; user rows come from the user module schema).
const (
	seederAssignRoleToUser = `INSERT INTO user_roles (
    user_id,
    role_id,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    current_timestamp,
    current_timestamp
) RETURNING
    user_role_id,
    user_id,
    role_id,
    created_at,
    updated_at,
    deleted_at`

	seederGetUserRoles = `SELECT
    r.role_id,
    r.role_name,
    r.created_at,
    r.updated_at,
    r.deleted_at
FROM
    roles r
JOIN
    user_roles ur ON ur.role_id = r.role_id
WHERE
    ur.user_id = $1
ORDER BY
    r.created_at ASC`

	seederUserRoleGetRoles = `SELECT
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

	seederGetUsers = `SELECT
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM users
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR firstname ILIKE '%' || $1 || '%'
        OR lastname ILIKE '%' || $1 || '%'
        OR email ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET $3`
)

// seedUser mirrors the user module's GetUsersRow for sqlx scanning (the user
// module's generated types carry no db tags).
type seedUser struct {
	UserID     int32            `db:"user_id"`
	Firstname  string           `db:"firstname"`
	Lastname   string           `db:"lastname"`
	Email      string           `db:"email"`
	Password   string           `db:"password"`
	CreatedAt  pgtype.Timestamp `db:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at"`
	TotalCount int64            `db:"total_count"`
}

// userRoleSeeder assigns seeded users to seeded roles. users, roles and
// user_roles all live in the identity context database (DB_IDENTITY); the two
// handles are kept for lifecycle symmetry with the other multi-handle seeders.
type userRoleSeeder struct {
	userDB *sqlx.DB
	roleDB *sqlx.DB
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewUserRoleSeeder(userDB *sqlx.DB, roleDB *sqlx.DB, ctx context.Context, logger logger.LoggerInterface) *userRoleSeeder {
	return &userRoleSeeder{
		userDB: userDB,
		roleDB: roleDB,
		ctx:    ctx,
		logger: logger,
	}
}

func (r *userRoleSeeder) Seed() error {
	var users []seedUser
	err := r.userDB.SelectContext(r.ctx, &users, seederGetUsers,
		"",
		int32(20),
		int32(0),
	)
	if err != nil {
		r.logger.Error("failed to fetch users", zap.Error(err))
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	var roles []db.GetRolesRow
	err = r.roleDB.SelectContext(r.ctx, &roles, seederUserRoleGetRoles,
		"",
		int32(4),
		int32(0),
	)
	if err != nil {
		r.logger.Error("failed to fetch roles", zap.Error(err))
		return fmt.Errorf("failed to fetch roles: %w", err)
	}

	if len(users) == 0 || len(roles) == 0 {
		r.logger.Debug("no users or roles available for seeding")
		return nil
	}

	// Idempotency: skip when every seeded user already has a role assigned.
	var assigned []db.Role
	err = r.roleDB.SelectContext(r.ctx, &assigned, seederGetUserRoles, users[0].UserID)
	if err == nil && len(assigned) > 0 {
		r.logger.Debug("user roles already seeded, skipping")
		return nil
	}

	rand.Seed(uint64(time.Now().UnixNano()))

	for _, user := range users {
		role := roles[rand.Intn(len(roles))]

		_, err := r.roleDB.ExecContext(r.ctx, seederAssignRoleToUser, user.UserID, role.RoleID)
		if err != nil {
			r.logger.Error("failed to assign role to user", zap.String("user", user.Email), zap.String("role", role.RoleName), zap.Error(err))
			return fmt.Errorf("failed to assign role %s to user %s: %w", role.RoleName, user.Email, err)
		}
	}

	r.logger.Info("user roles assigned successfully")
	return nil
}
