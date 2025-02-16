package handlers

import (
	cstErrors "github.com/ArtemSarafannikov/AvitoTestTask/internal/error"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/model"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/service"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/utils"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"
	"net/http"
)

type Handler struct {
	logger             echo.Logger
	userService        *service.UserService
	transactionService *service.TransactionService
}

func NewHandler(logger echo.Logger, userService *service.UserService, transactionService *service.TransactionService) *Handler {
	return &Handler{
		logger:             logger,
		userService:        userService,
		transactionService: transactionService,
	}
}

func (h *Handler) GetInfo(c echo.Context) error {
	userId := c.Get(utils.UserIdCtxKey).(string)
	ctx := c.Request().Context()

	var (
		coins     int
		inventory []*model.InfoInventory
		history   *model.CoinHistory
	)

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		var err error
		coins, err = h.userService.GetUserBalance(ctx, userId)
		return err
	})

	eg.Go(func() error {
		var err error
		inventory, err = h.transactionService.GetInventory(ctx, userId)
		return err
	})

	eg.Go(func() error {
		var err error
		history, err = h.transactionService.GetTransactionsHistory(ctx, userId)
		return err
	})

	if err := eg.Wait(); err != nil {
		cstErr := cstErrors.GetAndLogCustomError(err, h.logger)
		errResp := model.ErrorResponse{Errors: cstErr.Error()}
		return c.JSON(cstErr.(cstErrors.KnownError).Code(), errResp)
	}

	resp := model.InfoResponse{
		Balance:     coins,
		Inventory:   inventory,
		CoinHistory: history,
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) SendCoin(c echo.Context) error {
	var req model.SendCoinRequest
	if err := c.Bind(&req); err != nil {
		cstErr := cstErrors.GetAndLogCustomError(err, h.logger)
		errResp := model.ErrorResponse{Errors: cstErr.Error()}
		return c.JSON(cstErr.(cstErrors.KnownError).Code(), errResp)
	}
	userId := c.Get(utils.UserIdCtxKey).(string)
	if err := h.transactionService.SendCoin(c.Request().Context(), userId, req.ToUser, req.Amount); err != nil {
		cstErr := cstErrors.GetAndLogCustomError(err, h.logger)
		errResp := model.ErrorResponse{Errors: cstErr.Error()}
		return c.JSON(cstErr.(cstErrors.KnownError).Code(), errResp)
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handler) BuyItem(c echo.Context) error {
	itemId := c.Param("item")
	userId := c.Get(utils.UserIdCtxKey).(string)
	if err := h.transactionService.BuyItem(c.Request().Context(), userId, itemId); err != nil {
		cstErr := cstErrors.GetAndLogCustomError(err, h.logger)
		errResp := model.ErrorResponse{Errors: cstErr.Error()}
		return c.JSON(cstErr.(cstErrors.KnownError).Code(), errResp)
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handler) AuthHandler(c echo.Context) error {
	var req model.AuthRequest
	if err := c.Bind(&req); err != nil {
		cstErr := cstErrors.GetAndLogCustomError(err, h.logger)
		errResp := model.ErrorResponse{Errors: cstErr.Error()}
		return c.JSON(cstErr.(cstErrors.KnownError).Code(), errResp)
	}

	token, err := h.userService.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		cstErr := cstErrors.GetAndLogCustomError(err, h.logger)
		errResp := model.ErrorResponse{Errors: cstErr.Error()}
		return c.JSON(cstErr.(cstErrors.KnownError).Code(), errResp)
	}
	response := model.AuthResponse{Token: token}
	return c.JSON(http.StatusOK, response)
}
