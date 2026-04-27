package repository

import (
	"gorm.io/gorm"

	"github.com/edalferes/monetics/internal/modules/auth/adapters/repository/model"
	"github.com/edalferes/monetics/internal/modules/auth/domain"
	"github.com/edalferes/monetics/internal/modules/auth/usecase/interfaces"
)

// AuditLogRepository is a GORM-backed implementation of interfaces.AuditLogRepository.
type AuditLogRepository struct {
	DB *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{DB: db}
}

var _ interfaces.AuditLogRepository = (*AuditLogRepository)(nil)

func (r *AuditLogRepository) Create(log *domain.AuditLog) error {
	m := model.AuditLogFromDomain(log)
	if err := r.DB.Create(m).Error; err != nil {
		return err
	}
	log.ID = m.ID
	return nil
}

func (r *AuditLogRepository) ListAll() ([]domain.AuditLog, error) {
	var ms []model.AuditLogModel
	if err := r.DB.Order("created_at desc").Find(&ms).Error; err != nil {
		return nil, err
	}
	return model.AuditLogModelsToDomain(ms), nil
}
