package repository

import (
	"context"

	"github.com/vgartg/goauction/internal/models"
)

type LotRepository interface {
	CreateLot(ctx context.Context, lot *models.Lot) error
	GetLotByID(ctx context.Context, id string, forUpdate bool) (*models.Lot, error)
	UpdateLot(ctx context.Context, lot *models.Lot, oldVersion int) error
	CreateBid(ctx context.Context, bid *models.Bid) error
	GetHighestBid(ctx context.Context, lotID string) (*models.Bid, error)
	GetActiveLots(ctx context.Context) ([]*models.Lot, error)
	GetAllLots(ctx context.Context) ([]*models.Lot, error)
}
