package auction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/vgartg/goauction/internal/models"
	"github.com/vgartg/goauction/internal/repository"
)

type Engine struct {
	repo      repository.LotRepository
	wsManager interface {
		BroadcastToLot(lotID string, message interface{})
	}
}

func NewEngine(repo repository.LotRepository, wsManager interface {
	BroadcastToLot(lotID string, message interface{})
}) *Engine {
	return &Engine{
		repo:      repo,
		wsManager: wsManager,
	}
}

func (e *Engine) CreateLot(ctx context.Context, title string, startPrice, minStep float64, closingAt time.Time) (*models.Lot, error) {
	lot := &models.Lot{
		Title:        title,
		StartPrice:   startPrice,
		MinStep:      minStep,
		CurrentPrice: startPrice,
		Status:       models.LotStatusActive,
		ClosingAt:    closingAt,
		Version:      1,
	}
	if err := e.repo.CreateLot(ctx, lot); err != nil {
		return nil, err
	}
	go e.StartTimerForLot(lot)
	return lot, nil
}

func (e *Engine) GetLot(ctx context.Context, id string) (*models.Lot, error) {
	return e.repo.GetLotByID(ctx, id, false)
}

func (e *Engine) ListLots(ctx context.Context) ([]*models.Lot, error) {
	return e.repo.GetAllLots(ctx)
}

func (e *Engine) PlaceBid(ctx context.Context, lotID, userID string, amount float64) (*models.Lot, error) {
	for attempts := 0; attempts < 3; attempts++ {
		lot, err := e.repo.GetLotByID(ctx, lotID, true)
		if err != nil {
			return nil, err
		}
		if lot.Status != models.LotStatusActive {
			return nil, errors.New("lot is not active")
		}
		if time.Now().After(lot.ClosingAt) {
			return nil, errors.New("lot already closed")
		}
		if amount <= lot.CurrentPrice {
			return nil, fmt.Errorf("bid must be higher than current price %.2f", lot.CurrentPrice)
		}
		if amount < lot.CurrentPrice+lot.MinStep {
			return nil, fmt.Errorf("bid must be at least %.2f more than current price", lot.MinStep)
		}

		bid := &models.Bid{
			LotID:  lotID,
			UserID: userID,
			Amount: amount,
		}
		if err := e.repo.CreateBid(ctx, bid); err != nil {
			return nil, err
		}

		oldVersion := lot.Version
		lot.CurrentPrice = amount
		lot.Version++
		if err := e.repo.UpdateLot(ctx, lot, oldVersion); err != nil {
			if errors.Is(err, repository.ErrOptimisticLock) {
				continue
			}
			return nil, err
		}

		e.wsManager.BroadcastToLot(lotID, map[string]interface{}{
			"type":      "new_bid",
			"lot_id":    lotID,
			"user_id":   userID,
			"amount":    amount,
			"new_price": amount,
			"timestamp": time.Now(),
		})
		return lot, nil
	}
	return nil, errors.New("failed to place bid after retries")
}

func (e *Engine) CloseLot(lot *models.Lot) error {
	ctx := context.Background()
	if lot.Status != models.LotStatusActive {
		return nil
	}
	highestBid, err := e.repo.GetHighestBid(ctx, lot.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if highestBid != nil {
		lot.WinnerID = &highestBid.UserID
	}
	lot.Status = models.LotStatusClosed
	if err := e.repo.UpdateLot(ctx, lot, lot.Version); err != nil {
		return err
	}
	e.wsManager.BroadcastToLot(lot.ID, map[string]interface{}{
		"type":        "lot_closed",
		"lot_id":      lot.ID,
		"winner_id":   lot.WinnerID,
		"final_price": lot.CurrentPrice,
	})
	return nil
}

func (e *Engine) StartTimerForLot(lot *models.Lot) {
	duration := time.Until(lot.ClosingAt)
	if duration <= 0 {
		if err := e.CloseLot(lot); err != nil {
			slog.Error("failed to close lot", "lot_id", lot.ID, "error", err)
		}
		return
	}
	time.AfterFunc(duration, func() {
		ctx := context.Background()
		freshLot, err := e.repo.GetLotByID(ctx, lot.ID, false)
		if err != nil {
			slog.Error("failed to reload lot for closing", "lot_id", lot.ID, "error", err)
			return
		}
		if freshLot.Status == models.LotStatusActive {
			if err := e.CloseLot(freshLot); err != nil {
				slog.Error("failed to close lot", "lot_id", lot.ID, "error", err)
			}
		}
	})
}
