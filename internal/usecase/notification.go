package usecase

import (
	"fmt"
	"log"
	"time"

	"github.com/financial-planning/internal/domain"
)

type NotificationUseCase struct {
	repo       domain.NotificationRepository
	budgetRepo domain.BudgetRepository
}

func NewNotificationUseCase(repo domain.NotificationRepository, budgetRepo domain.BudgetRepository) *NotificationUseCase {
	return &NotificationUseCase{repo: repo, budgetRepo: budgetRepo}
}

func (uc *NotificationUseCase) GetNotifications(userID int, unreadOnly bool) ([]domain.Notification, int, error) {
	notifs, err := uc.repo.GetByUserID(userID, unreadOnly)
	if err != nil {
		log.Printf("notification: GetNotifications userID=%d: %v", userID, err)
		return nil, 0, err
	}
	count, err := uc.repo.GetUnreadCount(userID)
	if err != nil {
		log.Printf("notification: GetUnreadCount userID=%d: %v", userID, err)
		return nil, 0, err
	}
	return notifs, count, nil
}

func (uc *NotificationUseCase) MarkRead(id, userID int) error {
	if err := uc.repo.MarkRead(id, userID); err != nil {
		log.Printf("notification: MarkRead id=%d userID=%d: %v", id, userID, err)
		return err
	}
	return nil
}

func (uc *NotificationUseCase) MarkAllRead(userID int) error {
	if err := uc.repo.MarkAllRead(userID); err != nil {
		log.Printf("notification: MarkAllRead userID=%d: %v", userID, err)
		return err
	}
	return nil
}

func (uc *NotificationUseCase) Delete(id, userID int) error {
	if err := uc.repo.Delete(id, userID); err != nil {
		log.Printf("notification: Delete id=%d userID=%d: %v", id, userID, err)
		return err
	}
	return nil
}

func (uc *NotificationUseCase) GetPreferences(userID int) (*domain.NotificationPreferences, error) {
	prefs, err := uc.repo.GetPreferences(userID)
	if err != nil {
		log.Printf("notification: GetPreferences userID=%d: %v", userID, err)
	}
	return prefs, err
}

func (uc *NotificationUseCase) UpdatePreferences(userID int, prefs domain.NotificationPreferences) error {
	if err := uc.repo.UpsertPreferences(userID, prefs); err != nil {
		log.Printf("notification: UpdatePreferences userID=%d: %v", userID, err)
		return err
	}
	return nil
}

// CheckBudgetAlerts inspects current month budget usage and creates
// BUDGET_WARNING or BUDGET_EXCEEDED notifications for thresholds breached.
// Called after transaction mutations.
func (uc *NotificationUseCase) CheckBudgetAlerts(userID int) {
	now := time.Now()
	prefs, err := uc.repo.GetPreferences(userID)
	if err != nil {
		log.Printf("notification: CheckBudgetAlerts GetPreferences userID=%d: %v", userID, err)
		return
	}
	if !prefs.BudgetAlerts {
		return
	}

	usages, err := uc.budgetRepo.GetUsage(userID, int(now.Month()), now.Year())
	if err != nil {
		log.Printf("notification: CheckBudgetAlerts GetUsage userID=%d: %v", userID, err)
		return
	}

	for _, u := range usages {
		var notifType string
		switch u.Status {
		case "EXCEEDED":
			notifType = "BUDGET_EXCEEDED"
		case "WARNING":
			notifType = "BUDGET_WARNING"
		default:
			continue
		}

		exists, err := uc.repo.ExistsRecent(userID, notifType, u.BudgetID)
		if err != nil || exists {
			continue
		}

		entityType := "budget"
		entityID := u.BudgetID
		var title, message string
		if notifType == "BUDGET_EXCEEDED" {
			title = fmt.Sprintf("Budget exceeded: %s", u.Category)
			message = fmt.Sprintf("Your %s budget has been exceeded (%.0f%% used).", u.Category, u.Percentage)
		} else {
			title = fmt.Sprintf("Budget warning: %s", u.Category)
			message = fmt.Sprintf("Your %s budget is at %.0f%% — approaching the limit.", u.Category, u.Percentage)
		}
		_ = uc.repo.Create(userID, notifType, title, message, &entityType, &entityID)
	}
}
