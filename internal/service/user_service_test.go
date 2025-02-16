package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cstErrors "github.com/ArtemSarafannikov/AvitoTestTask/internal/error"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/model"
	"github.com/ArtemSarafannikov/AvitoTestTask/internal/utils"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	args := m.Called(ctx, login)
	if user := args.Get(0); user != nil {
		return user.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	args := m.Called(ctx, user)
	if u := args.Get(0); u != nil {
		return u.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) GetUserById(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if user := args.Get(0); user != nil {
		return user.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

// --- Tests for UserService.Login ---

func TestUserService_Login_BadRequest(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	token, err := svc.Login(ctx, "", "password")
	assert.Error(t, err)
	assert.Equal(t, cstErrors.BadRequestDataError, err)
	assert.Empty(t, token)

	token, err = svc.Login(ctx, "username", "")
	assert.Error(t, err)
	assert.Equal(t, cstErrors.BadRequestDataError, err)
	assert.Empty(t, token)
}

func TestUserService_Login_GetUserByLoginError(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	someErr := errors.New("db error")
	repo.On("GetUserByLogin", ctx, "username").Return(nil, someErr)

	token, err := svc.Login(ctx, "username", "password")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UserService.Login")
	assert.Empty(t, token)
	repo.AssertExpectations(t)
}

func TestUserService_Login_RegisterFlow_RegisterError(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	repo.On("GetUserByLogin", ctx, "newuser").Return(nil, cstErrors.NotFoundError)
	repo.On("CreateUser", ctx, mock.AnythingOfType("*model.User")).Return(nil, errors.New("creation error"))

	token, err := svc.Login(ctx, "newuser", "password")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "creation error")
	assert.Empty(t, token)
	repo.AssertExpectations(t)
}

func TestUserService_Login_RegisterFlow_Success(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	repo.On("GetUserByLogin", ctx, "newuser").Return(nil, cstErrors.NotFoundError)
	createdUser := &model.User{Id: "123", Username: "newuser", Password: "dummy_hashed", Balance: 1000}
	repo.On("CreateUser", ctx, mock.AnythingOfType("*model.User")).Return(createdUser, nil)

	token, err := svc.Login(ctx, "newuser", "password")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	repo.AssertExpectations(t)
}

func TestUserService_Login_BadCredential(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	hashed, err := utils.HashPassword("correct_password")
	assert.NoError(t, err)

	existingUser := &model.User{Id: "123", Username: "existing", Password: hashed}
	repo.On("GetUserByLogin", ctx, "existing").Return(existingUser, nil)

	token, err := svc.Login(ctx, "existing", "wrong_password")
	assert.Error(t, err)
	assert.Equal(t, cstErrors.BadCredentialError, err)
	assert.Empty(t, token)
	repo.AssertExpectations(t)
}

func TestUserService_Login_Success(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	hashed, err := utils.HashPassword("correct_password")
	assert.NoError(t, err)

	existingUser := &model.User{Id: "123", Username: "existing", Password: hashed}
	repo.On("GetUserByLogin", ctx, "existing").Return(existingUser, nil)

	token, err := svc.Login(ctx, "existing", "correct_password")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	repo.AssertExpectations(t)
}

// --- Tests for UserService.GetUserBalance ---

func TestUserService_GetUserBalance_CustomError(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	repo.On("GetUserById", ctx, "123").Return(nil, cstErrors.InternalError)

	balance, err := svc.GetUserBalance(ctx, "123")
	assert.Error(t, err)
	assert.Equal(t, 0, balance)
	repo.AssertExpectations(t)
}

func TestUserService_GetUserBalance_NonCustomError(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	nonCustomErr := errors.New("db error")
	repo.On("GetUserById", ctx, "123").Return(nil, nonCustomErr)

	balance, err := svc.GetUserBalance(ctx, "123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UserService.GetUserBalance")
	assert.Equal(t, 0, balance)
	repo.AssertExpectations(t)
}

func TestUserService_GetUserBalance_Success(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	user := &model.User{Id: "123", Balance: 1000}
	repo.On("GetUserById", ctx, "123").Return(user, nil)

	balance, err := svc.GetUserBalance(ctx, "123")
	assert.NoError(t, err)
	assert.Equal(t, 1000, balance)
	repo.AssertExpectations(t)
}
