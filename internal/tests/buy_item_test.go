package tests

import (
	"context"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/config"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/model"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/repository"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/service"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_BuyItem_Success(t *testing.T) {
	cfg := config.MustLoad()

	repo, err := repository.NewPostgresRepository(cfg.Storage)
	require.NoError(t, err)

	ts := service.NewTransactionService(repo)

	ctx := context.Background()

	hashedPassword, err := utils.HashPassword("test_password")
	user := &model.User{
		Username: "test_user",
		Password: hashedPassword,
		Balance:  1000,
	}
	user, err = repo.CreateUser(ctx, user)
	require.NoError(t, err)

	merch := &model.Merch{
		Name:      "Test Item",
		Price:     500,
		IsSelling: true,
	}
	merch, err = repo.CreateMerch(ctx, merch)
	require.NoError(t, err)

	err = ts.BuyItem(ctx, user.Id, merch.Id)
	require.NoError(t, err)

	updatedUser, err := repo.GetUserById(ctx, user.Id)
	require.NoError(t, err)
	assert.Equal(t, 500, updatedUser.Balance)

	inventory, err := repo.GetInventory(ctx, user.Id)
	require.NoError(t, err)
	assert.Len(t, inventory, 1)
	assert.Equal(t, merch.Name, inventory[0].Type)
}
