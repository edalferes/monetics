package auth

import (
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/edalferes/monetics/internal/config"
	"github.com/edalferes/monetics/internal/modules/auth/adapters/crypto"
	"github.com/edalferes/monetics/internal/modules/auth/adapters/http/handlers"
	"github.com/edalferes/monetics/internal/modules/auth/adapters/repository"
	"github.com/edalferes/monetics/internal/modules/auth/adapters/token"
	"github.com/edalferes/monetics/internal/modules/auth/usecase/audit"
	"github.com/edalferes/monetics/internal/modules/auth/usecase/interfaces"
	permUC "github.com/edalferes/monetics/internal/modules/auth/usecase/permission"
	roleUC "github.com/edalferes/monetics/internal/modules/auth/usecase/role"
	userUC "github.com/edalferes/monetics/internal/modules/auth/usecase/user"
	"github.com/edalferes/monetics/pkg/logger"
)

type Module struct {
	db        *gorm.DB
	jwtSecret string
	log       logger.Logger

	// Repositories
	userRepo     interfaces.User
	roleRepo     interfaces.Role
	permRepo     interfaces.PermissionRepository
	auditLogRepo interfaces.AuditLogRepository

	// Services
	passwordService interfaces.PasswordService
	jwtService      interfaces.JWTService
	auditService    interfaces.AuditService

	// Handlers
	publicHandler        *handlers.LoginHandler
	adminUserHandler     *handlers.AdminUserHandler
	adminRolePermHandler *handlers.AdminHandler
	auditHandler         *handlers.AuditHandler
	userHandler          *handlers.UserHandler
}

// NewModule constructs the auth module. The standard signature (db, cfg, log) keeps
// composition consistent across modules.
func NewModule(db *gorm.DB, cfg *config.Config, log logger.Logger) *Module {
	jwtSecret := cfg.JWT.Secret
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)

	passwordService := crypto.NewBcryptPasswordService()
	jwtService := token.NewJWTService(jwtSecret, time.Duration(cfg.JWT.ExpiryHour)*time.Hour)
	auditService := audit.NewAuditService(auditLogRepo)

	// Use Cases
	loginUC := userUC.NewLoginWithAuditUseCase(userRepo, passwordService, jwtService, auditService, log)

	listUsersUC := userUC.NewListUsersUseCase(userRepo, log)
	createUserUC := userUC.NewRegisterUseCase(userRepo, roleRepo, passwordService, log)
	getUserByIDUC := userUC.NewGetUserByIDUseCase(userRepo, log)
	updateUserUC := &userUC.UpdateUserByAdminUseCase{
		UserRepo:        userRepo,
		PasswordService: passwordService,
	}
	deleteUserUC := userUC.NewDeleteUserUseCase(userRepo, log)
	assignRoleUC := &userUC.AssignRoleUseCase{
		UserRepo: userRepo,
		RoleRepo: roleRepo,
	}
	removeRoleUC := &userUC.RemoveRoleUseCase{
		UserRepo: userRepo,
		RoleRepo: roleRepo,
	}
	changePasswordUC := userUC.NewChangePasswordUseCase(userRepo, passwordService, log)

	listRolesUC := &roleUC.ListRolesUseCase{RoleRepo: roleRepo}
	createRoleUC := &roleUC.CreateRoleUseCase{RoleRepo: roleRepo}
	deleteRoleUC := &roleUC.DeleteRoleUseCase{RoleRepo: roleRepo}

	listPermissionsUC := &permUC.ListPermissionsUseCase{PermissionRepo: permRepo}
	createPermissionUC := &permUC.CreatePermissionUseCase{PermissionRepo: permRepo}
	deletePermissionUC := &permUC.DeletePermissionUseCase{PermissionRepo: permRepo}

	// Handlers
	publicHandler := &handlers.LoginHandler{
		LoginUseCase: loginUC,
		Logger:       log,
	}

	adminUserHandler := &handlers.AdminUserHandler{
		ListUsersUC:   listUsersUC,
		CreateUserUC:  createUserUC,
		GetUserByIDUC: getUserByIDUC,
		UpdateUserUC:  updateUserUC,
		DeleteUserUC:  deleteUserUC,
		AssignRoleUC:  assignRoleUC,
		RemoveRoleUC:  removeRoleUC,
	}

	adminRolePermHandler := &handlers.AdminHandler{
		ListRolesUC:        listRolesUC,
		CreateRoleUC:       createRoleUC,
		DeleteRoleUC:       deleteRoleUC,
		ListPermissionsUC:  listPermissionsUC,
		CreatePermissionUC: createPermissionUC,
		DeletePermissionUC: deletePermissionUC,
	}

	auditHandler := handlers.NewAuditHandler(auditLogRepo)

	userHandler := &handlers.UserHandler{
		ChangePasswordUC: changePasswordUC,
	}

	return &Module{
		db:                   db,
		jwtSecret:            jwtSecret,
		log:                  log,
		userRepo:             userRepo,
		roleRepo:             roleRepo,
		permRepo:             permRepo,
		auditLogRepo:         auditLogRepo,
		passwordService:      passwordService,
		jwtService:           jwtService,
		auditService:         auditService,
		publicHandler:        publicHandler,
		adminUserHandler:     adminUserHandler,
		adminRolePermHandler: adminRolePermHandler,
		auditHandler:         auditHandler,
		userHandler:          userHandler,
	}
}

func (m *Module) RegisterRoutes(group *echo.Group) {
	group.POST("/auth/login", m.publicHandler.Login)

	adminGroup := group.Group("/admin")
	adminGroup.Use(JWTMiddleware(m.jwtSecret))
	adminGroup.Use(RequireRoles("admin"))

	adminGroup.POST("/users", m.adminUserHandler.CreateUser)
	adminGroup.GET("/users", m.adminUserHandler.ListUsers)
	adminGroup.GET("/users/:id", m.adminUserHandler.GetUser)
	adminGroup.PUT("/users/:id", m.adminUserHandler.UpdateUser)
	adminGroup.DELETE("/users/:id", m.adminUserHandler.DeleteUser)
	adminGroup.POST("/users/:id/roles", m.adminUserHandler.AssignRoleToUser)
	adminGroup.DELETE("/users/:id/roles/:roleName", m.adminUserHandler.RemoveRoleFromUser)

	adminGroup.GET("/roles", m.adminRolePermHandler.ListRoles)
	adminGroup.POST("/roles", m.adminRolePermHandler.CreateRole)
	adminGroup.DELETE("/roles/:name", m.adminRolePermHandler.DeleteRole)

	adminGroup.GET("/permissions", m.adminRolePermHandler.ListPermissions)
	adminGroup.POST("/permissions", m.adminRolePermHandler.CreatePermission)
	adminGroup.DELETE("/permissions/:name", m.adminRolePermHandler.DeletePermission)

	adminGroup.GET("/audit-logs", m.auditHandler.ListAuditLogs)

	userGroup := group.Group("/user")
	userGroup.Use(JWTMiddleware(m.jwtSecret))
	userGroup.PUT("/password", m.userHandler.ChangePassword)
}

// WireUp initializes and registers the auth module.
func WireUp(group *echo.Group, db *gorm.DB, cfg *config.Config, log logger.Logger) {
	module := NewModule(db, cfg, log)
	module.RegisterRoutes(group)
	log.Info().Msg("Auth module started successfully")
}
