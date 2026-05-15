package auction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/vgartg/goauction/internal/metrics"
	"github.com/vgartg/goauction/internal/models"
	"github.com/vgartg/goauction/internal/repository"
)

type WSBroadcaster interface {
	BroadcastToLot(lotID string, message interface{})
}

type Config struct {
	// Anti-sniping: if a bid lands within SnipingWindow before close,
	// extend ClosingAt by SnipingExtension. Zero window disables it.
	SnipingWindow    time.Duration
	SnipingExtension time.Duration
}

type Engine struct {
	repo repository.LotRepository
	ws   WSBroadcaster
	cfg  Config

	timersMu sync.Mutex
	timers   map[string]*time.Timer
}

func NewEngine(repo repository.LotRepository, ws WSBroadcaster, cfg Config) *Engine {
	return &Engine{
		repo:   repo,
		ws:     ws,
		cfg:    cfg,
		timers: make(map[string]*time.Timer),
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
	metrics.ActiveLots.Inc()
	e.scheduleClose(lot)
	return lot, nil
}

func (e *Engine) GetLot(ctx context.Context, id string) (*models.Lot, error) {
	return e.repo.GetLotByID(ctx, id, false)
}

func (e *Engine) ListLots(ctx context.Context) ([]*models.Lot, error) {
	return e.repo.GetAllLots(ctx)
}

func (e *Engine) RecentBids(ctx context.Context, lotID string, limit int) ([]*models.Bid, error) {
	return e.repo.GetRecentBids(ctx, lotID, limit)
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

		// Anti-sniping: extend close window if bid lands near the end.
		extended := false
		if e.cfg.SnipingWindow > 0 && time.Until(lot.ClosingAt) <= e.cfg.SnipingWindow {
			lot.ClosingAt = time.Now().Add(e.cfg.SnipingExtension)
			lot.ExtendedCount++
			extended = true
		}

		if err := e.repo.UpdateLot(ctx, lot, oldVersion); err != nil {
			if errors.Is(err, repository.ErrOptimisticLock) {
				continue
			}
			return nil, err
		}

		metrics.BidsTotal.WithLabelValues(lotID).Inc()

		e.ws.BroadcastToLot(lotID, map[string]interface{}{
			"type":      "new_bid",
			"lot_id":    lotID,
			"user_id":   userID,
			"amount":    amount,
			"new_price": amount,
			"timestamp": time.Now().UTC(),
		})

		if extended {
			e.scheduleClose(lot)
			e.ws.BroadcastToLot(lotID, map[string]interface{}{
				"type":           "lot_extended",
				"lot_id":         lotID,
				"closing_at":     lot.ClosingAt.UTC(),
				"extended_count": lot.ExtendedCount,
			})
		}
		return lot, nil
	}
	return nil, errors.New("failed to place bid after retries")
}

func (e *Engine) CloseLot(lot *models.Lot) error {
	ctx := context.Background()
	if lot.Status != models.LotStatusActive {
		return nil
	}
	if time.Now().Before(lot.ClosingAt) {
		// Closing was extended after this timer was scheduled. Reschedule and bail.
		e.scheduleClose(lot)
		return nil
	}
	highestBid, err := e.repo.GetHighestBid(ctx, lot.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if highestBid != nil {
		lot.WinnerID = &highestBid.UserID
	}
	oldVersion := lot.Version
	lot.Status = models.LotStatusClosed
	lot.Version++
	if err := e.repo.UpdateLot(ctx, lot, oldVersion); err != nil {
		return err
	}
	metrics.LotClosuresTotal.Inc()
	metrics.ActiveLots.Dec()
	e.clearTimer(lot.ID)
	e.ws.BroadcastToLot(lot.ID, map[string]interface{}{
		"type":        "lot_closed",
		"lot_id":      lot.ID,
		"winner_id":   lot.WinnerID,
		"final_price": lot.CurrentPrice,
	})
	return nil
}

// StartTimerForLot is used at boot to restore timers for active lots.
func (e *Engine) StartTimerForLot(lot *models.Lot) {
	e.scheduleClose(lot)
}

// scheduleClose installs (or replaces) the auto-close timer for a lot.
func (e *Engine) scheduleClose(lot *models.Lot) {
	duration := time.Until(lot.ClosingAt)
	if duration <= 0 {
		go func() {
			if err := e.closeLotByID(lot.ID); err != nil {
				slog.Error("failed to close lot", "lot_id", lot.ID, "error", err)
			}
		}()
		return
	}
	timer := time.AfterFunc(duration, func() {
		if err := e.closeLotByID(lot.ID); err != nil {
			slog.Error("failed to close lot", "lot_id", lot.ID, "error", err)
		}
	})
	e.timersMu.Lock()
	if old, ok := e.timers[lot.ID]; ok {
		old.Stop()
	}
	e.timers[lot.ID] = timer
	e.timersMu.Unlock()
}

func (e *Engine) clearTimer(lotID string) {
	e.timersMu.Lock()
	defer e.timersMu.Unlock()
	if t, ok := e.timers[lotID]; ok {
		t.Stop()
		delete(e.timers, lotID)
	}
}

func (e *Engine) closeLotByID(lotID string) error {
	ctx := context.Background()
	fresh, err := e.repo.GetLotByID(ctx, lotID, false)
	if err != nil {
		return err
	}
	if fresh.Status != models.LotStatusActive {
		return nil
	}
	return e.CloseLot(fresh)
}
