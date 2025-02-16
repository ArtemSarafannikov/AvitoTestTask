package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cstErrors "github.com/ArtemSarafannikov/AvitoTestTask/internal/error"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/model"
)

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) GetMerchById(ctx context.Context, itemId string) (*model.Merch, error) {
	args := m.Called(ctx, itemId)
	if merch := args.Get(0); merch != nil {
		return merch.(*model.Merch), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTransactionRepository) UpdateBalance(ctx context.Context, userId string, diffBalance int) error {
	args := m.Called(ctx, userId, diffBalance)
	return args.Error(0)
}

func (m *MockTransactionRepository) LogTransferCoin(ctx context.Context, fromUserId, toUserId string, amount int) error {
	args := m.Called(ctx, fromUserId, toUserId, amount)
	return args.Error(0)
}

func (m *MockTransactionRepository) LogBuyMerch(ctx context.Context, userId, merchId string, price int) error {
	args := m.Called(ctx, userId, merchId, price)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetTransactionHistoryReceived(ctx context.Context, userId string) ([]*model.ReceivedCoin, error) {
	args := m.Called(ctx, userId)
	if rec := args.Get(0); rec != nil {
		return rec.([]*model.ReceivedCoin), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTransactionRepository) GetTransactionHistorySent(ctx context.Context, userId string) ([]*model.SentCoin, error) {
	args := m.Called(ctx, userId)
	if sent := args.Get(0); sent != nil {
		return sent.([]*model.SentCoin), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTransactionRepository) GetInventory(ctx context.Context, userId string) ([]*model.InfoInventory, error) {
	args := m.Called(ctx, userId)
	if inv := args.Get(0); inv != nil {
		return inv.([]*model.InfoInventory), args.Error(1)
	}
	return nil, args.Error(1)
}

// --- Tests for TransactionService.SendCoin ---

func TestTransactionService_SendCoin_SameUser(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	ts := NewTransactionService(mockRepo)
	ctx := context.Background()

	err := ts.SendCoin(ctx, "user1", "user1", 100)
	assert.Error(t, err)
	assert.Equal(t, cstErrors.CantSendCoinYourselfError, err)
}

func TestTransactionService_SendCoin_UpdateBalanceFromError(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	ts := NewTransactionService(mockRepo)
	ctx := context.Background()

	normalErr := errors.New("update error")
	mockRepo.On("UpdateBalance", ctx, "user1", -100).Return(normalErr)

	err := ts.SendCoin(ctx, "user1", "user2", 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
	mockRepo.AssertExpectations(t)
}

func TestTransactionService_SendCoin_UpdateBalanceFromCustomError(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	ts := NewTransactionService(mockRepo)
	ctx := context.Background()

	customErr := cstErrors.InternalError
	mockRepo.On("UpdateBalance", ctx, "user1", -100).Return(customErr)

	err := ts.SendCoin(ctx, "user1", "user2", 100)
	assert.Error(t, err)
	assert.Equal(t, customErr, err)
	mockRepo.AssertExpectations(t)
}

func TestTransactionService_SendCoin_UpdateBalanceToCustomError(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	ts := NewTransactionService(mockRepo)
	ctx := context.Background()

	mockRepo.On("UpdateBalance", ctx, "user1", -100).Return(nil)
	customErr := cstErrors.InternalError
	mockRepo.On("UpdateBalance", ctx, "user2", 100).Return(customErr)

	err := ts.SendCoin(ctx, "user1", "user2", 100)
	assert.Error(t, err)
	assert.Equal(t, customErr, err)
	mockRepo.AssertExpectations(t)
}

func TestTransactionService_SendCoin_UpdateBalanceToError(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	ts := NewTransactionService(mockRepo)
	ctx := context.Background()

	mockRepo.On("UpdateBalance", ctx, "user1", -100).Return(nil)
	mockRepo.On("UpdateBalance", ctx, "user2", 100).Return(errors.New("update error to"))

	err := ts.SendCoin(ctx, "user1", "user2", 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error to")
	mockRepo.AssertExpectations(t)
}

func TestTransactionService_SendCoin_LogTransferError(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	ts := NewTransactionService(mockRepo)
	ctx := context.Background()

	mockRepo.On("UpdateBalance", ctx, "user1", -100).Return(nil)
	mockRepo.On("UpdateBalance", ctx, "user2", 100).Return(nil)
	normalErr := errors.New("log transfer error")
	mockRepo.On("LogTransferCoin", ctx, "user1", "user2", 100).Return(normalErr)

	err := ts.SendCoin(ctx, "user1", "user2", 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "log transfer error")
	mockRepo.AssertExpectations(t)
}

func TestTransactionService_SendCoin_Success(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	ts := NewTransactionService(mockRepo)
	ctx := context.Background()

	fromUserID := "user1"
	toUserID := "user2"
	amount := 50

	// Настраиваем мок репозитория
	mockRepo.On("UpdateBalance", ctx, fromUserID, -amount).Return(nil)
	mockRepo.On("UpdateBalance", ctx, toUserID, amount).Return(nil)
	mockRepo.On("LogTransferCoin", ctx, fromUserID, toUserID, amount).Return(nil)

	err := ts.SendCoin(ctx, fromUserID, toUserID, amount)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

// --- Tests for TransactionService.BuyItem ---

func TestTransactionService_BuyItem_GetMerchCustomError(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	ts := NewTransactionService(mockRepo)
	ctx := context.Background()

	customErr := cstErrors.InternalError
	mockRepo.On("GetMerchById", ctx, "item1").Return(nil, customErr)

	err := ts.BuyItem(ctx, "user1", "item1")
	assert.Error(t, err)
	assert.Equal(t, customErr, err)
	mockRepo.AssertExpectations(t)
}

func TestTransactionService_BuyItem_GetMerchError(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	ts := NewTransactionService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetMerchById", ctx, "item1").Return(nil, errors.New("merch error"))

	err := ts.BuyItem(ctx, "user1", "item1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "merch error")
	mockRepo.AssertExpectations(t)
}

func TestTransactionService_BuyItem_MerchNotSelling(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	ts := NewTransactionService(mockRepo)
	ctx := context.Background()

	merch := &model.Merch{Id: "item1", Price: 500, IsSelling: false}
	mockRepo.On("GetMerchById", ctx, "item1").Return(merch, nil)

	err := ts.BuyItem(ctx, "user1", "item1")
	assert.Error(t, err)
	assert.Equal(t, cstErrors.NoSellingMerchError, err)
	mockRepo.AssertExpectations(t)
}

func TestTransactionService_BuyItem_UpdateBalanceCustomError(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	ts := NewTransactionService(mockRepo)
	ctx := context.Background()

	merch := &model.Merch{Id: "item1", Price: 500, IsSelling: true}
	mockRepo.On("GetMerchById", ctx, "item1").Return(merch, nil)
	customErr := cstErrors.InternalError
	mockRepo.On("UpdateBalance", ctx, "user1", -500).Return(customErr)

	err := ts.BuyItem(ctx, "user1", "item1")
	assert.Error(t, err)
	assert.Equal(t, customErr, err)
	mockRepo.AssertExpectations(t)
}

func TestTransactionService_BuyItem_UpdateBalanceError(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	ts := NewTransactionService(mockRepo)
	ctx := context.Background()

	merch := &model.Merch{Id: "item1", Price: 500, IsSelling: true}
	mockRepo.On("GetMerchById", ctx, "item1").Return(merch, nil)
	mockRepo.On("UpdateBalance", ctx, "user1", -500).Return(errors.New("balance update error"))

	err := ts.BuyItem(ctx, "user1", "item1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "balance update error")
	mockRepo.AssertExpectations(t)
}

func TestTransactionService_BuyItem_LogBuyMerchError(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	ts := NewTransactionService(mockRepo)
	ctx := context.Background()

	merch := &model.Merch{Id: "item1", Price: 500, IsSelling: true}
	mockRepo.On("GetMerchById", ctx, "item1").Return(merch, nil)
	mockRepo.On("UpdateBalance", ctx, "user1", -500).Return(nil)
	normalErr := errors.New("log buy error")
	mockRepo.On("LogBuyMerch", ctx, "user1", "item1", 500).Return(normalErr)

	err := ts.BuyItem(ctx, "user1", "item1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "log buy error")
	mockRepo.AssertExpectations(t)
}

func TestTransactionService_BuyItem_Success(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	ts := NewTransactionService(mockRepo)
	ctx := context.Background()

	userID := "user1"
	itemID := "item1"
	price := 100

	merch := &model.Merch{Id: itemID, Price: price, IsSelling: true}

	mockRepo.On("GetMerchById", ctx, itemID).Return(merch, nil)
	mockRepo.On("UpdateBalance", ctx, userID, -price).Return(nil)
	mockRepo.On("LogBuyMerch", ctx, userID, itemID, price).Return(nil)

	err := ts.BuyItem(ctx, userID, itemID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// --- Tests for TransactionService.GetTransactionsHistory ---

func TestTransactionService_GetTransactionsHistory_ReceivedError(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	ts := NewTransactionService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetTransactionHistoryReceived", ctx, "user1").Return([]*model.ReceivedCoin(nil), errors.New("received error"))

	history, err := ts.GetTransactionsHistory(ctx, "user1")
	assert.Error(t, err)
	assert.Nil(t, history)
	mockRepo.AssertExpectations(t)
}

func TestTransactionService_GetTransactionsHistory_SentError(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	ts := NewTransactionService(mockRepo)
	ctx := context.Background()

	received := []*model.ReceivedCoin{{Amount: 50}} // имеются полученные транзакции
	mockRepo.On("GetTransactionHistoryReceived", ctx, "user1").Return(received, nil)
	mockRepo.On("GetTransactionHistorySent", ctx, "user1").Return([]*model.SentCoin(nil), errors.New("sent error"))

	history, err := ts.GetTransactionsHistory(ctx, "user1")
	assert.Error(t, err)
	assert.Nil(t, history)
	mockRepo.AssertExpectations(t)
}

func TestTransactionService_GetTransactionsHistory_Success(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	ts := NewTransactionService(mockRepo)
	ctx := context.Background()

	received := []*model.ReceivedCoin{{Amount: 50}}
	sent := []*model.SentCoin{{Amount: 30}}

	mockRepo.On("GetTransactionHistoryReceived", ctx, "user1").Return(received, nil)
	mockRepo.On("GetTransactionHistorySent", ctx, "user1").Return(sent, nil)

	history, err := ts.GetTransactionsHistory(ctx, "user1")
	assert.NoError(t, err)
	assert.Equal(t, received, history.Received)
	assert.Equal(t, sent, history.Sent)
	mockRepo.AssertExpectations(t)
}

// --- Tests for TransactionService.GetInventory ---

func TestTransactionService_GetInventory_Error(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	ts := NewTransactionService(mockRepo)
	ctx := context.Background()

	errMsg := "inventory error"
	mockRepo.On("GetInventory", ctx, "user1").Return(nil, errors.New(errMsg))

	inv, err := ts.GetInventory(ctx, "user1")
	assert.Error(t, err)
	assert.Nil(t, inv)
	assert.Contains(t, err.Error(), "TransactionService.GetInventory")
	mockRepo.AssertExpectations(t)
}

func TestTransactionService_GetInventory_Success(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	ts := NewTransactionService(mockRepo)
	ctx := context.Background()

	inventory := []*model.InfoInventory{{Type: "t-shirt", Quantity: 2}}
	mockRepo.On("GetInventory", ctx, "user1").Return(inventory, nil)

	result, err := ts.GetInventory(ctx, "user1")
	assert.NoError(t, err)
	assert.Equal(t, inventory, result)
	mockRepo.AssertExpectations(t)
}
