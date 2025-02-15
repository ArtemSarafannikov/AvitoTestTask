package service

import (
	"context"
	"fmt"
	cstErrors "github.com/ArtemSarafannikov/AvitoTestTask/internal/error"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/model"
)

type TransactionRepository interface {
	GetMerchById(ctx context.Context, itemId string) (*model.Merch, error)

	UpdateBalance(ctx context.Context, userId string, diffBalance int) error
	LogTransferCoin(ctx context.Context, fromUserId, toUserId string, amount int) error
	LogBuyMerch(ctx context.Context, userId, merchId string, price int) error
	GetTransactionHistoryReceived(ctx context.Context, userId string) ([]*model.ReceivedCoin, error)
	GetTransactionHistorySent(ctx context.Context, userId string) ([]*model.SentCoin, error)
	GetInventory(ctx context.Context, userId string) ([]*model.InfoInventory, error)
}

type TransactionService struct {
	repo TransactionRepository
}

func NewTransactionService(repo TransactionRepository) *TransactionService {
	return &TransactionService{
		repo: repo,
	}
}

func (t *TransactionService) SendCoin(ctx context.Context, fromUserId, toUserId string, amount int) error {
	const op = "TransactionService.SendCoin"

	if fromUserId == toUserId {
		return cstErrors.CantSendCoinYourselfError
	}

	var err error
	err = t.repo.UpdateBalance(ctx, fromUserId, -amount)
	if err != nil {
		if cstErrors.IsCustomError(err) {
			return err
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	err = t.repo.UpdateBalance(ctx, toUserId, amount)
	if err != nil {
		if cstErrors.IsCustomError(err) {
			return err
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	err = t.repo.LogTransferCoin(ctx, fromUserId, toUserId, amount)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (t *TransactionService) BuyItem(ctx context.Context, userId string, itemId string) error {
	const op = "TransactionService.BuyItem"

	var err error
	merch, err := t.repo.GetMerchById(ctx, itemId)
	if err != nil {
		if cstErrors.IsCustomError(err) {
			return err
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if !merch.IsSelling {
		return cstErrors.NoSellingMerchError
	}

	err = t.repo.UpdateBalance(ctx, userId, -merch.Price)
	if err != nil {
		if cstErrors.IsCustomError(err) {
			return err
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	err = t.repo.LogBuyMerch(ctx, userId, itemId, merch.Price)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (t *TransactionService) GetTransactionsHistory(ctx context.Context, userId string) (*model.CoinHistory, error) {
	const op = "TransactionService.GetTransactionsHistory"

	received, err := t.repo.GetTransactionHistoryReceived(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	sent, err := t.repo.GetTransactionHistorySent(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	coinHistory := &model.CoinHistory{
		Received: received,
		Sent:     sent,
	}
	return coinHistory, nil
}

func (t *TransactionService) GetInventory(ctx context.Context, userId string) ([]*model.InfoInventory, error) {
	const op = "TransactionService.GetInventory"

	inventory, err := t.repo.GetInventory(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return inventory, nil
}
