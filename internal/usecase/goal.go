package usecase

import (
	"errors"
	"log"
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
	goals, err := uc.repo.GetAll(userID, active)
	if err != nil {
		log.Printf("goal: GetGoals userID=%d active=%v: %v", userID, active, err)
	}
	return goals, err
}

func (uc *GoalUseCase) GetByID(id, userID int) (*domain.GoalResponse, error) {
	g, err := uc.repo.GetByID(id, userID)
	if err != nil {
		log.Printf("goal: GetByID id=%d userID=%d: %v", id, userID, err)
	}
	return g, err
}

func (uc *GoalUseCase) GetOverview(userID int) (*domain.GoalOverviewResponse, error) {
	savings, err := uc.repo.GetSavingsTotal(userID)
	if err != nil {
		log.Printf("goal: GetOverview GetSavingsTotal userID=%d: %v", userID, err)
		return nil, err
	}
	milestones, err := uc.repo.GetUpcomingMilestones(userID)
	if err != nil {
		log.Printf("goal: GetOverview GetUpcomingMilestones userID=%d: %v", userID, err)
		return nil, err
	}
	total, err := uc.repo.CountActive(userID)
	if err != nil {
		log.Printf("goal: GetOverview CountActive userID=%d: %v", userID, err)
		return nil, err
	}
	completedThisYear, err := uc.repo.CountCompletedThisYear(userID)
	if err != nil {
		log.Printf("goal: GetOverview CountCompletedThisYear userID=%d: %v", userID, err)
		return nil, err
	}
	return &domain.GoalOverviewResponse{
		TotalGoals:        total,
		Goals:             milestones,
		Savings:           int(savings),
		CompletedThisYear: completedThisYear,
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
	if err := uc.repo.Create(userID, req); err != nil {
		log.Printf("goal: Create userID=%d name=%s: %v", userID, req.Name, err)
		return err
	}
	return nil
}

func (uc *GoalUseCase) Update(id, userID int, req domain.CreateGoalRequest) error {
	if req.TargetAmount <= 0 {
		return errors.New("target amount must be greater than 0")
	}
	if req.Name == "" {
		return errors.New("name is required")
	}
	if err := uc.repo.Update(id, userID, req); err != nil {
		log.Printf("goal: Update id=%d userID=%d: %v", id, userID, err)
		return err
	}
	return nil
}

func (uc *GoalUseCase) Delete(id, userID int) error {
	if err := uc.repo.Delete(id, userID); err != nil {
		log.Printf("goal: Delete id=%d userID=%d: %v", id, userID, err)
		return err
	}
	return nil
}

func (uc *GoalUseCase) Contribute(id, userID, amount int) error {
	if amount <= 0 {
		return errors.New("contribution must be greater than 0")
	}
	net, err := uc.txRepo.GetNetSavings(userID)
	if err != nil {
		log.Printf("goal: Contribute GetNetSavings userID=%d: %v", userID, err)
		return err
	}
	if net <= 0 {
		return errors.New("cannot add contributions: no net savings")
	}
	if amount > int(net) {
		return errors.New("contribution exceeds available savings")
	}
	if err := uc.repo.Contribute(id, userID, amount); err != nil {
		log.Printf("goal: Contribute id=%d userID=%d amount=%d: %v", id, userID, amount, err)
		return err
	}
	return nil
}
