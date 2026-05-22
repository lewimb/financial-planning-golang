package usecase

import (
	"log"

	"github.com/financial-planning/internal/domain"
)

type ActivityLogUseCase struct {
	repo domain.ActivityLogRepository
}

func NewActivityLogUseCase(repo domain.ActivityLogRepository) *ActivityLogUseCase {
	return &ActivityLogUseCase{repo: repo}
}

func (uc *ActivityLogUseCase) GetActivity(userID int, limit, offset int) ([]domain.ActivityLog, int, error) {
	logs, total, err := uc.repo.GetByUserID(userID, limit, offset)
	if err != nil {
		log.Printf("activity_log: GetActivity userID=%d: %v", userID, err)
	}
	return logs, total, err
}
