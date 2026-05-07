package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/KV2013/gophermart-loyalty/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type UserRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

var userFieldsSql = `id, uuid, login, created_at, updated_at, deleted_at`

func NewUserRepository(db *sqlx.DB, logger *zap.Logger) (*UserRepository, error) {

	repo := &UserRepository{db: db, logger: logger}

	return repo, nil
}

func (r *UserRepository) FindByLogin(ctx context.Context, login string) (*model.User, error) {
	var user model.User
	err := r.db.GetContext(ctx, &user, `SELECT `+userFieldsSql+` FROM users WHERE login = $1 AND deleted_at IS NULL`, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("FindByLogin: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, login string, passwordHash string) (*model.User, error) {
	var user model.User

	newUserQuery := `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING ` + userFieldsSql

	row := r.db.QueryRowxContext(ctx, newUserQuery, login, passwordHash)

	if err := row.StructScan(&user); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, &model.ErrUserAlreadyExists{Login: login}
		}
		return nil, fmt.Errorf("Repository.Create: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) FindByCredentials(ctx context.Context, login string, passwordHash string) (*model.User, error) {
	var user model.User

	findByCredentialsQuery := `SELECT ` + userFieldsSql + ` FROM users WHERE login LIKE $1 AND password LIKE $2`
	err := r.db.GetContext(ctx, &user, findByCredentialsQuery, login, passwordHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &model.ErrUserNotFound{Login: login}
		}
		return nil, fmt.Errorf("FindByCredentials: %w", err)
	}

	if !user.DeletedAt.IsZero() {
		return nil, &model.ErrUserDeleted{Login: login, DeletedAt: user.DeletedAt}
	}

	return &user, nil
}
