package tests

import (
	"testing"

	"github.com/fishdivinity/BeeCount-Cloud/internal/auth"
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"github.com/fishdivinity/BeeCount-Cloud/internal/service"
	"github.com/fishdivinity/BeeCount-Cloud/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, utils.ErrUserNotFound
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, utils.ErrUserNotFound
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, utils.ErrUserNotFound
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) List(limit, offset int) ([]models.User, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]models.User), args.Error(1)
}

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) GenerateToken(user *models.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ValidateToken(tokenString string) (*auth.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, utils.ErrUnauthorized
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}

func (m *MockAuthService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) CheckPassword(password, hash string) error {
	args := m.Called(password, hash)
	return args.Error(0)
}

func TestUserService_Register(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockAuth := new(MockAuthService)
	userService := service.NewUserService(mockRepo, mockAuth)

	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "password123",
	}

	mockAuth.On("HashPassword", "password123").Return("hashedpassword", nil)
	mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)

	err := userService.Register(user)
	assert.NoError(t, err)

	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Login_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockAuth := new(MockAuthService)
	userService := service.NewUserService(mockRepo, mockAuth)

	user := &models.User{
		ID:           1,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}

	mockRepo.On("GetByEmail", "test@example.com").Return(user, nil)
	mockAuth.On("CheckPassword", "password123", "hashedpassword").Return(nil)
	mockAuth.On("GenerateToken", user).Return("token123", nil)

	_, token, err := userService.Login("test@example.com", "password123")
	assert.NoError(t, err)
	assert.Equal(t, "token123", token)

	mockRepo.AssertExpectations(t)
	mockAuth.AssertExpectations(t)
}

func TestUserService_Login_InvalidPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockAuth := new(MockAuthService)
	userService := service.NewUserService(mockRepo, mockAuth)

	user := &models.User{
		ID:           1,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}

	mockRepo.On("GetByEmail", "test@example.com").Return(user, nil)
	mockAuth.On("CheckPassword", "wrongpassword", "hashedpassword").Return(utils.ErrInvalidCredentials)

	_, _, err := userService.Login("test@example.com", "wrongpassword")
	assert.Error(t, err)
	assert.Equal(t, utils.ErrInvalidCredentials, err)

	mockRepo.AssertExpectations(t)
	mockAuth.AssertExpectations(t)
}

