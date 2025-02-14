package repository

import (
	"context"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/model"
)

type UserRepository interface {
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)

	CreateUser(ctx context.Context, user *model.User) (*model.User, error)
}

type TransactionRepository interface {
}

type MerchRepository interface {
}
