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
	_ "github.com/lib/pq"
	"github.com/vgartg/goauction/internal/models"
)

var ErrOptimisticLock = errors.New("optimistic lock failed")

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
              VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query,
		lot.Title, lot.StartPrice, lot.MinStep, lot.CurrentPrice, lot.Status, lot.ClosingAt, lot.Version,
	).Scan(&lot.ID, &lot.CreatedAt)
}

func (r *PostgresRepo) GetLotByID(ctx context.Context, id string, forUpdate bool) (*models.Lot, error) {
	query := `SELECT id, title, start_price, min_step, current_price, status, created_at, closing_at, version, winner_id
              FROM lots WHERE id = $1`
	if forUpdate {
		query += " FOR UPDATE"
	}
	row := r.db.QueryRowContext(ctx, query, id)
	lot := &models.Lot{}
	var winnerID sql.NullString
	err := row.Scan(&lot.ID, &lot.Title, &lot.StartPrice, &lot.MinStep, &lot.CurrentPrice,
		&lot.Status, &lot.CreatedAt, &lot.ClosingAt, &lot.Version, &winnerID)
	if err != nil {
		return nil, err
	}
	if winnerID.Valid {
		lot.WinnerID = &winnerID.String
	}
	return lot, nil
}

func (r *PostgresRepo) UpdateLot(ctx context.Context, lot *models.Lot, oldVersion int) error {
	query := `UPDATE lots SET current_price=$1, status=$2, version=$3, winner_id=$4
              WHERE id=$5 AND version=$6`
	result, err := r.db.ExecContext(ctx, query, lot.CurrentPrice, lot.Status, lot.Version,
		lot.WinnerID, lot.ID, oldVersion)
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
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, start_price, min_step, current_price, status, created_at, closing_at, version, winner_id
         FROM lots WHERE status = 'active' AND closing_at > NOW()`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lots []*models.Lot
	for rows.Next() {
		lot := &models.Lot{}
		var winnerID sql.NullString
		if err := rows.Scan(&lot.ID, &lot.Title, &lot.StartPrice, &lot.MinStep, &lot.CurrentPrice,
			&lot.Status, &lot.CreatedAt, &lot.ClosingAt, &lot.Version, &winnerID); err != nil {
			return nil, err
		}
		if winnerID.Valid {
			lot.WinnerID = &winnerID.String
		}
		lots = append(lots, lot)
	}
	return lots, nil
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
