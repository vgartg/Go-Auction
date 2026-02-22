package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
	"github.com/vgartg/goauction/internal/models"
)

var (
	ErrOptimisticLock = errors.New("optimistic lock failed")
	ErrUserExists     = errors.New("user already exists")
	ErrUserNotFound   = errors.New("user not found")
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(databaseURL string) (*PostgresRepo, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}
	return &PostgresRepo{db: db}, nil
}

func (r *PostgresRepo) Close() error {
	return r.db.Close()
}

func (r *PostgresRepo) RunMigrations(databaseURL string) error {
	driver, err := postgres.WithInstance(r.db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	slog.Info("migrations applied successfully")
	return nil
}

func (r *PostgresRepo) CreateLot(ctx context.Context, lot *models.Lot) error {
	query := `INSERT INTO lots (title, start_price, min_step, current_price, status, closing_at, version)
              VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, extended_count`
	return r.db.QueryRowContext(ctx, query,
		lot.Title, lot.StartPrice, lot.MinStep, lot.CurrentPrice, lot.Status, lot.ClosingAt, lot.Version,
	).Scan(&lot.ID, &lot.CreatedAt, &lot.ExtendedCount)
}

func (r *PostgresRepo) GetLotByID(ctx context.Context, id string, forUpdate bool) (*models.Lot, error) {
	query := `SELECT id, title, start_price, min_step, current_price, status, created_at, closing_at, version, winner_id, extended_count
              FROM lots WHERE id = $1`
	if forUpdate {
		query += " FOR UPDATE"
	}
	row := r.db.QueryRowContext(ctx, query, id)
	return scanLot(row)
}

func (r *PostgresRepo) UpdateLot(ctx context.Context, lot *models.Lot, oldVersion int) error {
	query := `UPDATE lots SET current_price=$1, status=$2, version=$3, winner_id=$4, closing_at=$5, extended_count=$6
              WHERE id=$7 AND version=$8`
	result, err := r.db.ExecContext(ctx, query,
		lot.CurrentPrice, lot.Status, lot.Version, lot.WinnerID, lot.ClosingAt, lot.ExtendedCount,
		lot.ID, oldVersion)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrOptimisticLock
	}
	return nil
}

func (r *PostgresRepo) GetActiveLots(ctx context.Context) ([]*models.Lot, error) {
	return r.queryLots(ctx,
		`SELECT id, title, start_price, min_step, current_price, status, created_at, closing_at, version, winner_id, extended_count
         FROM lots WHERE status = 'active' AND closing_at > NOW()`)
}

func (r *PostgresRepo) GetAllLots(ctx context.Context) ([]*models.Lot, error) {
	return r.queryLots(ctx, `
        SELECT id, title, start_price, min_step, current_price, status, created_at, closing_at, version, winner_id, extended_count
        FROM lots ORDER BY created_at DESC
    `)
}

func (r *PostgresRepo) CreateBid(ctx context.Context, bid *models.Bid) error {
	query := `INSERT INTO bids (lot_id, user_id, amount) VALUES ($1, $2, $3) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, bid.LotID, bid.UserID, bid.Amount).Scan(&bid.ID, &bid.CreatedAt)
}

func (r *PostgresRepo) GetHighestBid(ctx context.Context, lotID string) (*models.Bid, error) {
	query := `SELECT id, lot_id, user_id, amount, created_at FROM bids
              WHERE lot_id = $1 ORDER BY amount DESC LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, lotID)
	bid := &models.Bid{}
	err := row.Scan(&bid.ID, &bid.LotID, &bid.UserID, &bid.Amount, &bid.CreatedAt)
	if err != nil {
		return nil, err
	}
	return bid, nil
}

func (r *PostgresRepo) GetRecentBids(ctx context.Context, lotID string, limit int) ([]*models.Bid, error) {
	query := `SELECT id, lot_id, user_id, amount, created_at FROM bids
              WHERE lot_id = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, lotID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bids []*models.Bid
	for rows.Next() {
		b := &models.Bid{}
		if err := rows.Scan(&b.ID, &b.LotID, &b.UserID, &b.Amount, &b.CreatedAt); err != nil {
			return nil, err
		}
		bids = append(bids, b)
	}
	return bids, nil
}

func (r *PostgresRepo) CreateUser(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3)
              RETURNING id, created_at`
	err := r.db.QueryRowContext(ctx, query, user.Username, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrUserExists
		}
		return err
	}
	return nil
}

func (r *PostgresRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE email = $1`
	row := r.db.QueryRowContext(ctx, query, email)
	return scanUser(row)
}

func (r *PostgresRepo) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	return scanUser(row)
}

func (r *PostgresRepo) GetUserStats(ctx context.Context, id string) (*models.UserStats, error) {
	stats := &models.UserStats{UserID: id}
	query := `
        SELECT u.username,
               COALESCE((SELECT COUNT(*) FROM bids WHERE user_id = u.id), 0)            AS bids_count,
               COALESCE((SELECT COUNT(*) FROM lots WHERE winner_id = u.id), 0)          AS wins_count,
               COALESCE((SELECT SUM(current_price) FROM lots WHERE winner_id = u.id), 0) AS total_spent
        FROM users u
        WHERE u.id = $1`
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&stats.Username, &stats.BidsCount, &stats.WinsCount, &stats.TotalSpent)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *PostgresRepo) queryLots(ctx context.Context, query string) ([]*models.Lot, error) {
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lots []*models.Lot
	for rows.Next() {
		lot, err := scanLotRow(rows)
		if err != nil {
			return nil, err
		}
		lots = append(lots, lot)
	}
	return lots, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanLot(row rowScanner) (*models.Lot, error) {
	lot := &models.Lot{}
	var winnerID sql.NullString
	err := row.Scan(&lot.ID, &lot.Title, &lot.StartPrice, &lot.MinStep, &lot.CurrentPrice,
		&lot.Status, &lot.CreatedAt, &lot.ClosingAt, &lot.Version, &winnerID, &lot.ExtendedCount)
	if err != nil {
		return nil, err
	}
	if winnerID.Valid {
		lot.WinnerID = &winnerID.String
	}
	return lot, nil
}

func scanLotRow(rows *sql.Rows) (*models.Lot, error) {
	lot := &models.Lot{}
	var winnerID sql.NullString
	if err := rows.Scan(&lot.ID, &lot.Title, &lot.StartPrice, &lot.MinStep, &lot.CurrentPrice,
		&lot.Status, &lot.CreatedAt, &lot.ClosingAt, &lot.Version, &winnerID, &lot.ExtendedCount); err != nil {
		return nil, err
	}
	if winnerID.Valid {
		lot.WinnerID = &winnerID.String
	}
	return lot, nil
}

func scanUser(row rowScanner) (*models.User, error) {
	u := &models.User{}
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}
