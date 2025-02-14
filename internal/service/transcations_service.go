package service

import (
	"context"
	"fmt"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/repository"
)

type TransactionService struct {
	repo repository.TransactionRepository
}

func NewTransactionService(repo repository.TransactionRepository) *TransactionService {
	return &TransactionService{
		repo: repo,
	}
}

func (t *TransactionService) SendCoin(ctx context.Context, fromUserId, toUserId string, amount int) error {
	const op = "TransactionService.SendCoin"
	err := t.repo.TransferCoin(ctx, fromUserId, toUserId, amount)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (t *TransactionService) BuyItem() error {
	return nil
}
