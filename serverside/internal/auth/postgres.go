package auth

import (
	"authbackend/generated/sqlc"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type PostgresRepository struct{ q *sqlc.Queries }

func NewPostgresRepository(q *sqlc.Queries) *PostgresRepository { return &PostgresRepository{q} }

func (r *PostgresRepository) Create(ctx context.Context, u User) error {
	id, e := uuid.Parse(u.ID)
	if e != nil {
		return e
	}
	role := u.PlatformRole
	if role == "" {
		role = "user"
	}
	_, e = r.q.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           pgtype.UUID{Bytes: id, Valid: true},
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		PlatformRole: role,
	})
	var pe *pgconn.PgError
	if errors.As(e, &pe) && pe.Code == "23505" {
		return ErrUserExists
	}
	return e
}

func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (User, error) {
	u, e := r.q.GetUserByEmail(ctx, email)
	if errors.Is(e, pgx.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if e != nil {
		return User{}, e
	}
	return User{
		ID:           uuid.UUID(u.ID.Bytes).String(),
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		PlatformRole: u.PlatformRole,
		CreatedAt:    u.CreatedAt.Time,
		UpdatedAt:    u.UpdatedAt.Time,
	}, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (User, error) {
	uid, e := uuid.Parse(id)
	if e != nil {
		return User{}, ErrUserNotFound
	}
	u, e := r.q.GetUserByID(ctx, pgtype.UUID{Bytes: uid, Valid: true})
	if errors.Is(e, pgx.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if e != nil {
		return User{}, e
	}
	return User{
		ID:           uuid.UUID(u.ID.Bytes).String(),
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		PlatformRole: u.PlatformRole,
		CreatedAt:    u.CreatedAt.Time,
		UpdatedAt:    u.UpdatedAt.Time,
	}, nil
}

func (r *PostgresRepository) UpdatePassword(ctx context.Context, id, hash string) error {
	uid, e := uuid.Parse(id)
	if e != nil {
		return e
	}
	return r.q.UpdatePassword(ctx, sqlc.UpdatePasswordParams{ID: pgtype.UUID{Bytes: uid, Valid: true}, PasswordHash: hash})
}
