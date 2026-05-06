package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/KV2013/gophermart-loyalty/internal/model"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
)

type SQLXRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

var userFieldsSql = `id, uuid, login, created_at, updated_at, deleted_at`

func NewRepository(dsn string, logger *zap.Logger) (*SQLXRepository, error) {
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	repo := &SQLXRepository{db: db, logger: logger}

	err = repo.db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Проверка подключения не прошла: %w", err)
	}
	logger.Debug("Успешное подключение к БД")

	if err := repo.runMigrations(); err != nil {
		return nil, fmt.Errorf("ошибка выполнения миграций: %w", err)
	}

	return repo, nil
}

func (r *SQLXRepository) runMigrations() error {
	driver, err := migratepgx.WithInstance(r.db.DB, &migratepgx.Config{})
	if err != nil {
		return fmt.Errorf("не удалось создать драйвер миграций: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("не удалось инициализировать migrate: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("не удалось применить миграции: %w", err)
	}

	return nil
}

func (r *SQLXRepository) FindByLogin(ctx context.Context, login string) (*model.User, error) {
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

func (r *SQLXRepository) Create(ctx context.Context, login string, passwordHash string) (*model.User, error) {
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

func (r *SQLXRepository) FindByCredentials(ctx context.Context, login string, passwordHash string) (*model.User, error) {
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
