package tests

import (
	"testing"

	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"github.com/fishdivinity/BeeCount-Cloud/internal/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&models.User{})
	assert.NoError(t, err)

	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}

	err := repo.Create(user)
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}
	repo.Create(user)

	found, err := repo.GetByEmail("test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, user.Email, found.Email)
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}
	repo.Create(user)

	found, err := repo.GetByID(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
}

