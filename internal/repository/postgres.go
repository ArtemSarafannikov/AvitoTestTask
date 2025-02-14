package repository

import (
	"context"
	"database/sql"
	"fmt"
	cstErrors "github.com/ArtemSarafannikov/AvitoTestTask/internal/error"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/model"
	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository() (*PostgresRepository, error) {
	conn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		"postgres", "postgres", "localhost:5432", "db_market", "disable")
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	const op = "postgres.GetUserByLogin"
	const query = `SELECT id, login, password, balance, created_at
					FROM users WHERE login = $1`

	var user model.User

	row := r.db.QueryRowContext(ctx, query, login)

	if err := row.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, row.Err())
	}
	err := row.Scan(&user.Id,
		&user.Username,
		&user.Password,
		&user.Balance,
		&user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cstErrors.NotFoundError
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &user, nil
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	const op = "postgres.CreateUser"
	const query = `INSERT INTO users (login, password, balance)
					VALUES ($1, $2, $3)
					RETURNING id, created_at`

	row := r.db.QueryRowContext(ctx, query, user.Username, user.Password, user.Balance)
	if err := row.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err := row.Scan(&user.Id,
		&user.CreatedAt); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

func (r *PostgresRepository) TransferCoin(ctx context.Context, fromUserId, toUserId string, amount int) error {
	const op = "postgres.TransferCoin"
	const query = `UPDATE users
					SET balance = balance - $1
					WHERE id = $2;

					UPDATE users
					SET balance = balance + $1
					WHERE id = $3;

					INSERT INTO transactions(from_user_id, to_user_id, amount)
					VALUES ($2, $3, $1);`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if _, err = stmt.ExecContext(ctx, amount, fromUserId, toUserId); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
