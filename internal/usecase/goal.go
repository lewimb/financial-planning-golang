package usecase

import (
	"errors"
	"time"

	"github.com/financial-planning/internal/domain"
)

type GoalUseCase struct {
	repo   domain.GoalRepository
	txRepo domain.TransactionRepository
}

func NewGoalUseCase(repo domain.GoalRepository, txRepo domain.TransactionRepository) *GoalUseCase {
	return &GoalUseCase{repo: repo, txRepo: txRepo}
}

func (uc *GoalUseCase) GetGoals(userID int, active bool) ([]domain.GoalResponse, error) {
	return uc.repo.GetAll(userID, active)
}

func (uc *GoalUseCase) GetByID(id, userID int) (*domain.GoalResponse, error) {
	return uc.repo.GetByID(id, userID)
}

func (uc *GoalUseCase) GetOverview(userID int) (*domain.GoalOverviewResponse, error) {
	savings, err := uc.repo.GetSavingsTotal(userID)
	if err != nil {
		return nil, err
	}
	milestones, err := uc.repo.GetUpcomingMilestones(userID)
	if err != nil {
		return nil, err
	}
	total, err := uc.repo.CountActive(userID)
	if err != nil {
		return nil, err
	}
	return &domain.GoalOverviewResponse{
		TotalGoals: total,
		Goals:      milestones,
		Savings:    int(savings), // savings values are monetary amounts; truncation is acceptable
	}, nil
}

func (uc *GoalUseCase) Create(userID int, req domain.CreateGoalRequest) error {
	if req.TargetAmount <= 0 {
		return errors.New("target amount must be greater than 0")
	}
	if req.Deadline.Before(time.Now()) {
		return errors.New("deadline must be in the future")
	}
	if req.Name == "" {
		return errors.New("name is required")
	}
	return uc.repo.Create(userID, req)
}

func (uc *GoalUseCase) Update(id, userID int, req domain.CreateGoalRequest) error {
	if req.TargetAmount <= 0 {
		return errors.New("target amount must be greater than 0")
	}
	if req.Name == "" {
		return errors.New("name is required")
	}
	return uc.repo.Update(id, userID, req)
}

func (uc *GoalUseCase) Delete(id, userID int) error {
	return uc.repo.Delete(id, userID)
}

func (uc *GoalUseCase) Contribute(id, userID, amount int) error {
	if amount <= 0 {
		return errors.New("contribution must be greater than 0")
	}
	net, err := uc.txRepo.GetNetSavings(userID)
	if err != nil {
		return err
	}
	if net <= 0 {
		return errors.New("cannot add contributions: no net savings")
	}
	if amount > int(net) {
		return errors.New("contribution exceeds available savings")
	}
	return uc.repo.Contribute(id, userID, amount)
}
