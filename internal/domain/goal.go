package domain

import "time"

type GoalResponse struct {
	Id            int       `json:"id"`
	Name          string    `json:"name"`
	TargetAmount  int       `json:"target_amount"`
	CurrentAmount int       `json:"current_amount"`
	Status        string    `json:"status"`
	Deadline      time.Time `json:"deadline"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateGoalRequest struct {
	Name         string    `json:"name"`
	TargetAmount int       `json:"target_amount"`
	Description  string    `json:"description"`
	Deadline     time.Time `json:"deadline"`
}

type GoalContributionRequest struct {
	GoalId       int `json:"goal_id"`
	Contribution int `json:"contribution"`
}

type GoalOverviewResponse struct {
	TotalGoals int            `json:"total_goals"`
	Savings    int            `json:"savings"`
	Goals      []GoalResponse `json:"goals"`
}

type GoalRepository interface {
	GetAll(userID int, active bool) ([]GoalResponse, error)
	GetByID(id, userID int) (*GoalResponse, error)
	GetSavingsTotal(userID int) (float64, error)
	CountActive(userID int) (int, error)
	GetUpcomingMilestones(userID int) ([]GoalResponse, error)
	Create(userID int, req CreateGoalRequest) error
	Update(id, userID int, req CreateGoalRequest) error
	Delete(id, userID int) error
	Contribute(id, userID, amount int) error
}
