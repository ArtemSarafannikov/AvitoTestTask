package service

import (
	"context"
	"errors"
	"fmt"
	cstErrors "github.com/ArtemSarafannikov/AvitoTestTask/internal/error"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/model"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/utils"
)

type UserRepository interface {
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)

	CreateUser(ctx context.Context, user *model.User) (*model.User, error)
	GetUserById(ctx context.Context, id string) (*model.User, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (u *UserService) Login(ctx context.Context, username, password string) (string, error) {
	const op = "UserService.Login"
	if username == "" || password == "" {
		return "", cstErrors.BadRequestDataError
	}

	user, err := u.repo.GetUserByLogin(ctx, username)
	notFoundErr := errors.Is(err, cstErrors.NotFoundError)
	if err != nil && !notFoundErr {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if notFoundErr {
		// If user not exists
		user = &model.User{
			Username: username,
			Password: password,
		}
		user, err = u.Register(ctx, user)
		if err != nil {
			return "", fmt.Errorf("%s: %w", op, err)
		}
	} else {
		// If user exists
		if !utils.CheckPasswordHash(password, user.Password) {
			return "", cstErrors.BadCredentialError
		}
	}
	jwt, err := utils.GenerateJWT(user.Id)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return jwt, nil
}

func (u *UserService) Register(ctx context.Context, user *model.User) (*model.User, error) {
	const op = "UserService.Register"
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	user = &model.User{
		Username: user.Username,
		Password: hashedPassword,
		Balance:  1000,
	}
	user, err = u.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

func (u *UserService) GetUserBalance(ctx context.Context, userId string) (int, error) {
	const op = "UserService.GetUserBalance"
	user, err := u.repo.GetUserById(ctx, userId)
	if err != nil {
		if cstErrors.IsCustomError(err) {
			return 0, err
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return user.Balance, nil
}
