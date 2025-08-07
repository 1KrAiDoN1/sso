package postgres

import (
	"context"
	"errors"
	"fmt"
	"sso/internal/domain/models"
	"sso/internal/storage"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StoragePool struct {
	pool *pgxpool.Pool
}

func NewDatabase(ctx context.Context, databaseURL string) (*StoragePool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database connection string: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	// Проверка соединения
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &StoragePool{
			pool: pool},
		nil

}

func (d *StoragePool) GetPool() *pgxpool.Pool {
	return d.pool
}
func (d *StoragePool) Close() error {
	if d.pool != nil {
		d.pool.Close()
	}
	return nil
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

func (s *StoragePool) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "storage.postgres.SaveUser"

	query := `INSERT INTO users(email, pass_hash) VALUES($1, $2) RETURNING id`

	var id int64
	err := s.pool.QueryRow(ctx, query, email, passHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *StoragePool) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgres.User"

	query := `SELECT id, email, pass_hash FROM users WHERE email = $1`

	var user models.User
	err := s.pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *StoragePool) App(ctx context.Context, id int) (models.App, error) {
	const op = "storage.postgres.App"

	query := `SELECT id, name, secret FROM apps WHERE id = $1`

	var app models.App
	err := s.pool.QueryRow(ctx, query, id).Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}

func (s *StoragePool) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.postgres.IsAdmin"

	query := `SELECT is_admin FROM users WHERE id = $1`

	var isAdmin bool
	err := s.pool.QueryRow(ctx, query, userID).Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}
