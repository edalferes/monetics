package user_test

import (
	"testing"

	"github.com/edalferes/monetics/pkg/logger"

	"github.com/stretchr/testify/assert"

	"github.com/edalferes/monetics/internal/modules/auth/domain"
	"github.com/edalferes/monetics/internal/modules/auth/errors"
	"github.com/edalferes/monetics/internal/modules/auth/usecase/user"
)

func TestRegisterUseCase_Execute(t *testing.T) {
	t.Run("success - new user registration", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockRoleRepo := new(MockRoleRepository)
		mockPasswordService := new(MockPasswordService)
		uc := user.NewRegisterUseCase(mockUserRepo, mockRoleRepo, mockPasswordService, logger.NewDefault())
		_ = uc
		if false {
		}

		username := "newuser"
		password := "Password123"
		hashedPassword := "$2a$10$hashedpassword"

		// Mock user not exists
		mockUserRepo.On("FindByUsername", username).Return(nil, errors.ErrUserNotFound)

		// Mock password hashing
		mockPasswordService.On("Hash", password).Return(hashedPassword, nil)

		// Mock role exists
		userRole := &domain.Role{ID: 1, Name: "user"}
		mockRoleRepo.On("FindByName", "user").Return(userRole, nil)

		// Mock user creation
		expectedUser := &domain.User{
			Username: username,
			Password: hashedPassword,
			Roles:    []domain.Role{*userRole},
		}
		mockUserRepo.On("Create", expectedUser).Return(nil)

		err := uc.Execute(username, password)

		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
		mockRoleRepo.AssertExpectations(t)
		mockPasswordService.AssertExpectations(t)
	})

	t.Run("error - user already exists", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockRoleRepo := new(MockRoleRepository)
		mockPasswordService := new(MockPasswordService)
		uc := user.NewRegisterUseCase(mockUserRepo, mockRoleRepo, mockPasswordService, logger.NewDefault())
		_ = uc
		if false {
		}

		username := "existinguser"
		password := "Password123"

		// Mock user already exists
		existingUser := &domain.User{ID: 1, Username: username}
		mockUserRepo.On("FindByUsername", username).Return(existingUser, nil)

		err := uc.Execute(username, password)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserAlreadyExists, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("error - role not found", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockRoleRepo := new(MockRoleRepository)
		mockPasswordService := new(MockPasswordService)
		uc := user.NewRegisterUseCase(mockUserRepo, mockRoleRepo, mockPasswordService, logger.NewDefault())
		_ = uc
		if false {
		}

		username := "newuser"
		password := "Password123"
		hashedPassword := "$2a$10$hashedpassword"

		// Mock user not exists
		mockUserRepo.On("FindByUsername", username).Return(nil, errors.ErrUserNotFound)

		// Mock password hashing
		mockPasswordService.On("Hash", password).Return(hashedPassword, nil)

		// Mock role not found
		mockRoleRepo.On("FindByName", "user").Return(nil, errors.ErrUserNotFound)

		err := uc.Execute(username, password)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)
		mockUserRepo.AssertExpectations(t)
		mockRoleRepo.AssertExpectations(t)
		mockPasswordService.AssertExpectations(t)
	})
}

func TestListUsersUseCase_Execute(t *testing.T) {
	t.Run("success - list all users", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		uc := user.NewListUsersUseCase(mockUserRepo, logger.NewDefault())

		expectedUsers := []domain.User{
			{ID: 1, Username: "user1"},
			{ID: 2, Username: "user2"},
		}

		mockUserRepo.On("ListAll").Return(expectedUsers, nil)

		result, err := uc.Execute()

		assert.NoError(t, err)
		assert.Equal(t, expectedUsers, result)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("error - repository error", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		uc := user.NewListUsersUseCase(mockUserRepo, logger.NewDefault())

		expectedError := errors.ErrInvalidData
		mockUserRepo.On("ListAll").Return(nil, expectedError)

		result, err := uc.Execute()

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestGetUserByIDUseCase_Execute(t *testing.T) {
	t.Run("success - find user by id", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		uc := user.NewGetUserByIDUseCase(mockUserRepo, logger.NewDefault())

		userID := uint(1)
		expectedUser := &domain.User{ID: userID, Username: "testuser"}

		mockUserRepo.On("FindByID", userID).Return(expectedUser, nil)

		result, err := uc.Execute(userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, result)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("error - user not found", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		uc := user.NewGetUserByIDUseCase(mockUserRepo, logger.NewDefault())

		userID := uint(999)
		mockUserRepo.On("FindByID", userID).Return(nil, errors.ErrUserNotFound)

		result, err := uc.Execute(userID)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)
		assert.Nil(t, result)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUpdateUserUseCase_Execute(t *testing.T) {
	t.Run("success - update username only", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockRoleRepo := new(MockRoleRepository)
		uc := user.UpdateUserUseCase{
			User:       mockUserRepo,
			RoleReader: mockRoleRepo,
		}

		input := user.UpdateUserInput{
			ID:       1,
			Username: "newusername",
		}

		existingUser := &domain.User{
			ID:       1,
			Username: "oldusername",
			Password: "hashedpassword",
			Roles:    []domain.Role{{ID: 1, Name: "user"}},
		}

		mockUserRepo.On("FindByID", input.ID).Return(existingUser, nil)

		updatedUser := &domain.User{
			ID:       1,
			Username: "newusername",
			Password: "hashedpassword",
			Roles:    []domain.Role{{ID: 1, Name: "user"}},
		}
		mockUserRepo.On("Update", updatedUser).Return(nil)

		err := uc.Execute(input)

		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("success - update with roles", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockRoleRepo := new(MockRoleRepository)
		uc := user.UpdateUserUseCase{
			User:       mockUserRepo,
			RoleReader: mockRoleRepo,
		}

		input := user.UpdateUserInput{
			ID:      1,
			RoleIDs: []uint{1, 2},
		}

		existingUser := &domain.User{
			ID:       1,
			Username: "testuser",
			Password: "hashedpassword",
			Roles:    []domain.Role{},
		}

		role1 := &domain.Role{ID: 1, Name: "user"}
		role2 := &domain.Role{ID: 2, Name: "admin"}

		mockUserRepo.On("FindByID", input.ID).Return(existingUser, nil)
		mockRoleRepo.On("FindByID", uint(1)).Return(role1, nil)
		mockRoleRepo.On("FindByID", uint(2)).Return(role2, nil)

		updatedUser := &domain.User{
			ID:       1,
			Username: "testuser",
			Password: "hashedpassword",
			Roles:    []domain.Role{*role1, *role2},
		}
		mockUserRepo.On("Update", updatedUser).Return(nil)

		err := uc.Execute(input)

		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
		mockRoleRepo.AssertExpectations(t)
	})

	t.Run("error - user not found", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockRoleRepo := new(MockRoleRepository)
		uc := user.UpdateUserUseCase{
			User:       mockUserRepo,
			RoleReader: mockRoleRepo,
		}

		input := user.UpdateUserInput{
			ID:       999,
			Username: "newusername",
		}

		mockUserRepo.On("FindByID", input.ID).Return(nil, errors.ErrUserNotFound)

		err := uc.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestDeleteUserUseCase_Execute(t *testing.T) {
	t.Run("success - delete user", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		uc := user.NewDeleteUserUseCase(mockUserRepo, logger.NewDefault())

		userID := uint(1)
		mockUserRepo.On("Delete", userID).Return(nil)

		err := uc.Execute(userID)

		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("error - repository error", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		uc := user.NewDeleteUserUseCase(mockUserRepo, logger.NewDefault())

		userID := uint(1)
		expectedError := errors.ErrInvalidData
		mockUserRepo.On("Delete", userID).Return(expectedError)

		err := uc.Execute(userID)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockUserRepo.AssertExpectations(t)
	})
}
