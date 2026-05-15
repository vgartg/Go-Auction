package auction

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vgartg/goauction/internal/models"
	"github.com/vgartg/goauction/internal/repository"
)

type MockLotRepository struct {
	mock.Mock
}

func (m *MockLotRepository) CreateLot(ctx context.Context, lot *models.Lot) error {
	args := m.Called(ctx, lot)
	return args.Error(0)
}
func (m *MockLotRepository) GetLotByID(ctx context.Context, id string, forUpdate bool) (*models.Lot, error) {
	args := m.Called(ctx, id, forUpdate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Lot), args.Error(1)
}
func (m *MockLotRepository) UpdateLot(ctx context.Context, lot *models.Lot, oldVersion int) error {
	args := m.Called(ctx, lot, oldVersion)
	return args.Error(0)
}
func (m *MockLotRepository) CreateBid(ctx context.Context, bid *models.Bid) error {
	args := m.Called(ctx, bid)
	return args.Error(0)
}
func (m *MockLotRepository) GetHighestBid(ctx context.Context, lotID string) (*models.Bid, error) {
	args := m.Called(ctx, lotID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Bid), args.Error(1)
}
func (m *MockLotRepository) GetActiveLots(ctx context.Context) ([]*models.Lot, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Lot), args.Error(1)
}
func (m *MockLotRepository) GetAllLots(ctx context.Context) ([]*models.Lot, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Lot), args.Error(1)
}
func (m *MockLotRepository) GetRecentBids(ctx context.Context, lotID string, limit int) ([]*models.Bid, error) {
	args := m.Called(ctx, lotID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Bid), args.Error(1)
}

type MockWSManager struct {
	mock.Mock
}

func (m *MockWSManager) BroadcastToLot(lotID string, message interface{}) {
	m.Called(lotID, message)
}

func newTestEngine(repo *MockLotRepository, ws *MockWSManager) *Engine {
	return NewEngine(repo, ws, Config{
		SnipingWindow:    30 * time.Second,
		SnipingExtension: 30 * time.Second,
	})
}

func TestEngine_CreateLot(t *testing.T) {
	repo := new(MockLotRepository)
	ws := new(MockWSManager)
	engine := newTestEngine(repo, ws)

	ctx := context.Background()
	closingAt := time.Now().Add(10 * time.Minute)

	repo.On("CreateLot", ctx, mock.AnythingOfType("*models.Lot")).Return(nil).Run(func(args mock.Arguments) {
		lot := args.Get(1).(*models.Lot)
		lot.ID = "generated-id"
	})

	lot, err := engine.CreateLot(ctx, "Test", 100, 10, closingAt)
	assert.NoError(t, err)
	assert.Equal(t, "generated-id", lot.ID)
	repo.AssertExpectations(t)
}

func TestEngine_PlaceBid_RejectsBelowMinStep(t *testing.T) {
	repo := new(MockLotRepository)
	ws := new(MockWSManager)
	engine := newTestEngine(repo, ws)

	ctx := context.Background()
	existing := &models.Lot{
		ID:           "lot-1",
		Status:       models.LotStatusActive,
		CurrentPrice: 100,
		MinStep:      10,
		ClosingAt:    time.Now().Add(10 * time.Minute),
		Version:      1,
	}
	repo.On("GetLotByID", ctx, "lot-1", true).Return(existing, nil)

	_, err := engine.PlaceBid(ctx, "lot-1", "user-1", 105) // 100 + 10 = 110 required
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestEngine_PlaceBid_TriggersAntiSniping(t *testing.T) {
	repo := new(MockLotRepository)
	ws := new(MockWSManager)
	engine := newTestEngine(repo, ws)

	ctx := context.Background()
	existing := &models.Lot{
		ID:           "lot-1",
		Status:       models.LotStatusActive,
		CurrentPrice: 100,
		MinStep:      10,
		ClosingAt:    time.Now().Add(5 * time.Second), // within 30s sniping window
		Version:      1,
	}
	repo.On("GetLotByID", ctx, "lot-1", true).Return(existing, nil)
	repo.On("CreateBid", ctx, mock.AnythingOfType("*models.Bid")).Return(nil)
	repo.On("UpdateLot", ctx, mock.AnythingOfType("*models.Lot"), 1).Return(nil)

	ws.On("BroadcastToLot", "lot-1", mock.MatchedBy(func(m interface{}) bool {
		msg, ok := m.(map[string]interface{})
		return ok && msg["type"] == "new_bid"
	})).Return()
	ws.On("BroadcastToLot", "lot-1", mock.MatchedBy(func(m interface{}) bool {
		msg, ok := m.(map[string]interface{})
		return ok && msg["type"] == "lot_extended"
	})).Return()

	lot, err := engine.PlaceBid(ctx, "lot-1", "user-1", 110)
	assert.NoError(t, err)
	assert.Equal(t, 1, lot.ExtendedCount)
	assert.True(t, lot.ClosingAt.After(time.Now().Add(20*time.Second)))
	ws.AssertExpectations(t)
}

func TestEngine_PlaceBid_OptimisticLockRetry(t *testing.T) {
	repo := new(MockLotRepository)
	ws := new(MockWSManager)
	engine := newTestEngine(repo, ws)

	ctx := context.Background()
	// First fetch sees price=100, v=1. Concurrent bid bumps to price=105, v=2.
	// Retry fetches the new state and our 120 bid wins on v=2.
	repo.On("GetLotByID", ctx, "lot-1", true).Return(&models.Lot{
		ID: "lot-1", Status: models.LotStatusActive,
		CurrentPrice: 100, MinStep: 10,
		ClosingAt: time.Now().Add(10 * time.Minute), Version: 1,
	}, nil).Once()
	repo.On("GetLotByID", ctx, "lot-1", true).Return(&models.Lot{
		ID: "lot-1", Status: models.LotStatusActive,
		CurrentPrice: 105, MinStep: 10,
		ClosingAt: time.Now().Add(10 * time.Minute), Version: 2,
	}, nil).Once()

	repo.On("CreateBid", ctx, mock.AnythingOfType("*models.Bid")).Return(nil)

	repo.On("UpdateLot", ctx, mock.AnythingOfType("*models.Lot"), 1).
		Return(repository.ErrOptimisticLock).Once()
	repo.On("UpdateLot", ctx, mock.AnythingOfType("*models.Lot"), 2).
		Return(nil).Once()

	ws.On("BroadcastToLot", "lot-1", mock.Anything).Return()

	_, err := engine.PlaceBid(ctx, "lot-1", "user-1", 120)
	assert.NoError(t, err)
}

func TestEngine_PlaceBid_RejectsClosedLot(t *testing.T) {
	repo := new(MockLotRepository)
	ws := new(MockWSManager)
	engine := newTestEngine(repo, ws)

	ctx := context.Background()
	repo.On("GetLotByID", ctx, "lot-1", true).Return(&models.Lot{
		ID: "lot-1", Status: models.LotStatusClosed,
	}, nil)

	_, err := engine.PlaceBid(ctx, "lot-1", "user-1", 200)
	assert.True(t, errors.Is(err, errors.New("lot is not active")) || err.Error() == "lot is not active")
}
