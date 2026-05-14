package repository

import (
    "database/sql"
    "fmt"
    "log/slog"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
    _ "github.com/lib/pq"
    "github.com/vgartg/goauction/internal/models"
)

type PostgresRepo struct {
    db *sql.DB
}

func NewPostgresRepo(databaseURL string) (*PostgresRepo, error) {
    db, err := sql.Open("postgres", databaseURL)
    if err != nil {
        return nil, fmt.Errorf("failed to open db: %w", err)
    }
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

// Placeholder methods – to be implemented in commit 3
func (r *PostgresRepo) CreateLot(lot *models.Lot) error {
    return fmt.Errorf("not implemented")
}
func (r *PostgresRepo) GetLotByID(id string) (*models.Lot, error) {
    return nil, fmt.Errorf("not implemented")
}
func (r *PostgresRepo) UpdateLot(lot *models.Lot) error {
    return fmt.Errorf("not implemented")
}
func (r *PostgresRepo) CreateBid(bid *models.Bid) error {
    return fmt.Errorf("not implemented")
}