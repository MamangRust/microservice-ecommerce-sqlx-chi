package seeder

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-ecommerce-pkg/hash"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const getSeederUsers = `SELECT
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
OFFSET
    $3`

const createUserSeed = `INSERT INTO
    users (
        firstname,
        lastname,
        email,
        password,
        verification_code,
        is_verified,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        current_timestamp,
        current_timestamp
    )
RETURNING
    user_id`

const trashUserSeed = `UPDATE users
SET
    deleted_at = current_timestamp
WHERE
    user_id = $1
    AND deleted_at IS NULL`

type userSeeder struct {
	db     *sqlx.DB
	hash   hash.HashPassword
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewUserSeeder(db *sqlx.DB, hash hash.HashPassword, ctx context.Context, logger logger.LoggerInterface) *userSeeder {
	return &userSeeder{
		db:     db,
		hash:   hash,
		ctx:    ctx,
		logger: logger,
	}
}

func (r *userSeeder) Seed() error {
	// Idempotency: skip when users already exist.
	existing := []struct {
		UserID int32 `db:"user_id"`
	}{}
	err := r.db.SelectContext(r.ctx, &existing, getSeederUsers,
		"",
		int32(1),
		int32(0),
	)
	if err == nil && len(existing) > 0 {
		r.logger.Debug("users already seeded, skipping")
		return nil
	}

	for i := 1; i <= 10; i++ {
		email := fmt.Sprintf("user_%s@example.com", uuid.NewString())
		rawPassword := fmt.Sprintf("password%d", i)

		hashedPassword, err := r.hash.HashPassword(rawPassword)
		if err != nil {
			r.logger.Error("failed to hash password", zap.Int("user", i), zap.Error(err))
			return fmt.Errorf("failed to hash password for user %d: %w", i, err)
		}

		var createdUserID int32
		err = r.db.GetContext(r.ctx, &createdUserID, createUserSeed,
			fmt.Sprintf("User%d", i),
			fmt.Sprintf("Last%d", i),
			email,
			hashedPassword,
			"",
			false,
		)
		if err != nil {
			r.logger.Error("failed to seed user", zap.Int("user", i), zap.Error(err))
			return fmt.Errorf("failed to seed user %d: %w", i, err)
		}

		if i > 5 {
			_, err = r.db.ExecContext(r.ctx, trashUserSeed, createdUserID)
			if err != nil {
				r.logger.Error("failed to trash user", zap.Int("user", i), zap.Error(err))
				return fmt.Errorf("failed to trash user %d: %w", i, err)
			}
		}
	}

	r.logger.Info("User seeding completed successfully")
	return nil
}
