package auction

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vgartg/goauction/internal/models"
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

type MockWSManager struct {
	mock.Mock
}

func (m *MockWSManager) BroadcastToLot(lotID string, message interface{}) {
	m.Called(lotID, message)
}

func TestEngine_CreateLot(t *testing.T) {
	repo := new(MockLotRepository)
	ws := new(MockWSManager)
	engine := NewEngine(repo, ws)

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
