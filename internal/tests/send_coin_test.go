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

func Test_SendCoin_Success(t *testing.T) {
	cfg := config.MustLoad()

	var err error

	repo, err := repository.NewPostgresRepository(cfg.Storage)
	require.NoError(t, err)

	ts := service.NewTransactionService(repo)

	ctx := context.Background()

	var hashedPassword string
	hashedPassword, err = utils.HashPassword("sender_password")
	sender := &model.User{Username: "sender", Password: hashedPassword, Balance: 1000}

	hashedPassword, err = utils.HashPassword("reciever_password")
	receiver := &model.User{Username: "receiver", Password: hashedPassword, Balance: 200}

	sender, err = repo.CreateUser(ctx, sender)
	require.NoError(t, err)

	receiver, err = repo.CreateUser(ctx, receiver)
	require.NoError(t, err)

	err = ts.SendCoin(ctx, sender.Id, receiver.Id, 300)
	require.NoError(t, err)

	updatedSender, err := repo.GetUserById(ctx, sender.Id)
	require.NoError(t, err)
	assert.Equal(t, 700, updatedSender.Balance)

	updatedReceiver, err := repo.GetUserById(ctx, receiver.Id)
	require.NoError(t, err)
	assert.Equal(t, 500, updatedReceiver.Balance)

	history, err := repo.GetTransactionHistorySent(ctx, sender.Id)
	require.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, receiver.Username, history[0].ToUser)
	assert.Equal(t, 300, history[0].Amount)
}
