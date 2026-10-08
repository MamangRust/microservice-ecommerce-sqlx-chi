// Package user adapts the User query/command services gRPC API into the shared
// domain model. It is the only place that talks to pb/user.
package user

import (
	"context"
	"time"

	pbuser "github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/database/models"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/user_errors"
)

// QueryRepository is the contract consumers depend on for user reads.
type QueryRepository interface {
	FindByID(ctx context.Context, id int) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByEmailAndVerify(ctx context.Context, email string) (*models.User, error)
	FindByVerificationCode(ctx context.Context, code string) (*models.User, error)
}

// CommandRepository is the contract consumers depend on for user writes on
// behalf of the auth service.
type CommandRepository interface {
	CreateUser(ctx context.Context, request *requests.RegisterRequest) (*models.User, error)
	UpdateUserIsVerified(ctx context.Context, userID int, isVerified bool) (*models.User, error)
	UpdateUserPassword(ctx context.Context, userID int, password string) (*models.User, error)
	DeleteUserPermanent(ctx context.Context, userID int) error
}

// Repository implements QueryRepository and CommandRepository on top of the
// generated user query/command service clients.
type Repository struct {
	queryClient   pbuser.UserQueryServiceClient
	commandClient pbuser.UserCommandServiceClient
	guard         *resilience.DependencyGuard
}

// NewQueryAdapter wraps a generated user query client into a QueryRepository.
// Zero or more guard options attach a per-call timeout, bulkhead and circuit
// breaker around every outbound gRPC call; with no options the guard stays nil
// and calls pass straight through.
func NewQueryAdapter(client pbuser.UserQueryServiceClient, opts ...adapter.GuardOption) QueryRepository {
	r := &Repository{queryClient: client}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// NewAdapter wraps generated user query/command clients into a Repository that
// satisfies both QueryRepository and CommandRepository.
func NewAdapter(queryClient pbuser.UserQueryServiceClient, commandClient pbuser.UserCommandServiceClient, opts ...adapter.GuardOption) *Repository {
	r := &Repository{queryClient: queryClient, commandClient: commandClient}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Repository) SetGuard(g *resilience.DependencyGuard) {
	r.guard = g
}

func (r *Repository) FindByID(ctx context.Context, id int) (*models.User, error) {
	var res *pbuser.ApiResponseUser
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = r.queryClient.FindById(ctx, &pbuser.FindByIdUserRequest{Id: int32(id)})
		return callErr
	})
	if err != nil || res == nil || res.Data == nil {
		return nil, user_errors.ErrUserNotFound.WithInternal(err)
	}
	return userToModel(res.Data), nil
}

// FindByEmail returns the user with the password hash so callers can verify
// credentials. A missing user comes back as ErrUserNotFound.
func (r *Repository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var res *pbuser.ApiResponseUserWithPassword
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = r.queryClient.FindByEmail(ctx, &pbuser.FindByEmailRequest{Email: email})
		return callErr
	})
	if err != nil || res == nil || res.Data == nil {
		return nil, user_errors.ErrUserNotFound.WithInternal(err)
	}
	return userWithPasswordToModel(res.Data), nil
}

// FindByEmailAndVerify mirrors the legacy auth repo behaviour used on login.
// The user service exposes a single FindByEmail lookup, so this delegates to it.
func (r *Repository) FindByEmailAndVerify(ctx context.Context, email string) (*models.User, error) {
	return r.FindByEmail(ctx, email)
}

func (r *Repository) FindByVerificationCode(ctx context.Context, code string) (*models.User, error) {
	var res *pbuser.ApiResponseUser
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = r.queryClient.FindByVerificationCode(ctx, &pbuser.FindByVerificationCodeRequest{VerificationCode: code})
		return callErr
	})
	if err != nil || res == nil || res.Data == nil {
		return nil, user_errors.ErrUserNotFound.WithInternal(err)
	}
	return userToModel(res.Data), nil
}

// CreateUser registers a user through the User service. The password must
// already be hashed by the caller.
func (r *Repository) CreateUser(ctx context.Context, request *requests.RegisterRequest) (*models.User, error) {
	var res *pbuser.ApiResponseUser
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = r.commandClient.Create(ctx, &pbuser.CreateUserRequest{
			Firstname:       request.FirstName,
			Lastname:        request.LastName,
			Email:           request.Email,
			Password:        request.Password,
			ConfirmPassword: request.ConfirmPassword,
		})
		return callErr
	})
	if err != nil || res == nil || res.Data == nil {
		return nil, user_errors.ErrCreateUser.WithInternal(err)
	}
	return userToModel(res.Data), nil
}

func (r *Repository) UpdateUserIsVerified(ctx context.Context, userID int, isVerified bool) (*models.User, error) {
	var res *pbuser.ApiResponseUser
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = r.commandClient.UpdateIsVerified(ctx, &pbuser.UpdateUserIsVerifiedRequest{
			Id:         int32(userID),
			IsVerified: isVerified,
		})
		return callErr
	})
	if err != nil || res == nil || res.Data == nil {
		return nil, user_errors.ErrUpdateUserVerificationCode.WithInternal(err)
	}
	return userToModel(res.Data), nil
}

func (r *Repository) UpdateUserPassword(ctx context.Context, userID int, password string) (*models.User, error) {
	var res *pbuser.ApiResponseUser
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = r.commandClient.UpdatePassword(ctx, &pbuser.UpdateUserPasswordRequest{
			Id:       int32(userID),
			Password: password,
		})
		return callErr
	})
	if err != nil || res == nil || res.Data == nil {
		return nil, user_errors.ErrUpdateUserPassword.WithInternal(err)
	}
	return userToModel(res.Data), nil
}

// DeleteUserPermanent removes a user for good. The User service only purges rows
// that are already soft-deleted, so this first trashes the user and then purges
// it. Callers (auth's register compensation) just want the user gone.
func (r *Repository) DeleteUserPermanent(ctx context.Context, userID int) error {
	// Best effort: a user that is already trashed returns an error here, which
	// must not stop the purge below.
	_ = r.guard.Call(ctx, func(ctx context.Context) error {
		_, callErr := r.commandClient.TrashedUser(ctx, &pbuser.FindByIdUserRequest{Id: int32(userID)})
		return callErr
	})

	var res *pbuser.ApiResponseUserDelete
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = r.commandClient.DeleteUserPermanent(ctx, &pbuser.FindByIdUserRequest{Id: int32(userID)})
		return callErr
	})
	if err != nil || res == nil {
		return user_errors.ErrDeleteUserPermanent.WithInternal(err)
	}
	return nil
}

func userToModel(u *pbuser.UserResponse) *models.User {
	if u == nil {
		return nil
	}
	return &models.User{
		UserID:    u.Id,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Email:     u.Email,
		CreatedAt: parseTimeString(u.CreatedAt),
		UpdatedAt: parseTimeString(u.UpdatedAt),
	}
}

func userWithPasswordToModel(u *pbuser.UserResponseWithPassword) *models.User {
	if u == nil {
		return nil
	}
	return &models.User{
		UserID:    u.Id,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: parseTimeString(u.CreatedAt),
		UpdatedAt: parseTimeString(u.UpdatedAt),
	}
}

func parseTimeString(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02 15:04:05.000", s)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05Z", s)
		if err != nil {
			return nil
		}
	}
	return &t
}
