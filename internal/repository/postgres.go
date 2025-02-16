package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/config"
	cstErrors "github.com/ArtemSarafannikov/AvitoTestTask/internal/error"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/model"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(config config.DatabaseConfig) (*PostgresRepository, error) {
	conn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		config.User, config.Password, config.Address, config.Name, config.SSLMode)
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) isCheckConstraintViolation(err error) bool {
	if pgErr, ok := err.(*pq.Error); ok {
		return pgErr.Code == "23514"
	}
	return false
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

func (r *PostgresRepository) GetUserById(ctx context.Context, id string) (*model.User, error) {
	const op = "postgres.GetUserById"
	const query = `SELECT login, password, balance, created_at
					FROM users WHERE id = $1`

	var user model.User
	row := r.db.QueryRowContext(ctx, query, id)
	if err := row.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err := row.Scan(&user.Username,
		&user.Password,
		&user.Balance,
		&user.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, cstErrors.NotFoundError
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	user.Id = id
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

func (r *PostgresRepository) UpdateBalance(ctx context.Context, userId string, diffBalance int) error {
	const op = "postgres.UpdateBalance"
	const query = `UPDATE users
					SET balance = balance + $1
					WHERE id = $2;`
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if _, err = stmt.ExecContext(ctx, diffBalance, userId); err != nil {
		if r.isCheckConstraintViolation(err) {
			return cstErrors.NoCoinError
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *PostgresRepository) LogTransferCoin(ctx context.Context, fromUserId, toUserId string, amount int) error {
	const op = "postgres.TransferCoin"
	const query = `INSERT INTO transactions(from_user_id, to_user_id, amount)
					VALUES ($1, $2, $3);`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if _, err = stmt.ExecContext(ctx, fromUserId, toUserId, amount); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *PostgresRepository) GetMerchById(ctx context.Context, itemId string) (*model.Merch, error) {
	const op = "postgres.GetMerchById"
	const query = `SELECT name, price, is_selling, created_at
					FROM merch WHERE id = $1`

	var merch model.Merch

	row := r.db.QueryRowContext(ctx, query, itemId)
	if err := row.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err := row.Scan(&merch.Name,
		&merch.Price,
		&merch.IsSelling,
		&merch.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, cstErrors.NotFoundError
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	merch.Id = itemId
	return &merch, nil
}

func (r *PostgresRepository) LogBuyMerch(ctx context.Context, userId, merchId string, price int) error {
	const op = "postgres.LogBuyMerch"
	const query = `INSERT INTO purchases(user_id, merch_id, price)
					VALUES ($1, $2, $3);`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if _, err = stmt.ExecContext(ctx, userId, merchId, price); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *PostgresRepository) GetTransactionHistoryReceived(ctx context.Context, userId string) ([]*model.ReceivedCoin, error) {
	const op = "postgres.GetTransactionHistoryReceived"
	const query = `SELECT u.login from_user, amount
					FROM transactions t
					LEFT JOIN users u on u.id = t.from_user_id
					WHERE to_user_id = $1
					ORDER BY t.created_at DESC`

	var err error
	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var transactions []*model.ReceivedCoin
	for rows.Next() {
		var (
			t        model.ReceivedCoin
			fromUser sql.NullString
		)
		if err = rows.Scan(&fromUser, &t.Amount); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		if fromUser.Valid {
			t.FromUser = fromUser.String
		} else {
			t.FromUser = "DELETED USER"
		}
		transactions = append(transactions, &t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return transactions, nil
}

func (r *PostgresRepository) GetTransactionHistorySent(ctx context.Context, userId string) ([]*model.SentCoin, error) {
	const op = "postgres.GetTransactionHistoryReceived"
	const query = `SELECT u.login to_user, amount
					FROM transactions t
					LEFT JOIN users u on u.id = t.to_user_id
					WHERE from_user_id = $1
					ORDER BY t.created_at DESC`

	var err error
	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var transactions []*model.SentCoin
	for rows.Next() {
		var (
			t      model.SentCoin
			toUser sql.NullString
		)

		if err = rows.Scan(&toUser, &t.Amount); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		if toUser.Valid {
			t.ToUser = toUser.String
		} else {
			t.ToUser = "DELETED USER"
		}
		transactions = append(transactions, &t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return transactions, nil
}

func (r *PostgresRepository) GetInventory(ctx context.Context, userId string) ([]*model.InfoInventory, error) {
	const op = "postgres.GetInventory"
	const query = `SELECT m.name, COUNT(*)
					FROM purchases p
					LEFT JOIN merch m on p.merch_id = m.id
					WHERE user_id = $1
					GROUP BY m.name`

	var err error
	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var inventory []*model.InfoInventory
	for rows.Next() {
		var i model.InfoInventory
		if err = rows.Scan(&i.Type, &i.Quantity); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		inventory = append(inventory, &i)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return inventory, nil
}
