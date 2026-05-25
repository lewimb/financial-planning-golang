package delivery

import (
	"github.com/financial-planning/internal/delivery/http/handler"
	"github.com/financial-planning/internal/delivery/http/middleware"
	"github.com/financial-planning/internal/usecase"
	"github.com/gin-gonic/gin"
)

type Deps struct {
	UserUC             *usecase.UserUseCase
	TransactionUC      *usecase.TransactionUseCase
	BudgetUC           *usecase.BudgetUseCase
	GoalUC             *usecase.GoalUseCase
	DashboardUC        *usecase.DashboardUseCase
	ChatUC             *usecase.ChatUseCase
	MLUC               *usecase.MLUseCase
	FinancialProfileUC *usecase.FinancialProfileUseCase
	NotificationUC     *usecase.NotificationUseCase
	ActivityLogUC      *usecase.ActivityLogUseCase
	ReportsUC          *usecase.ReportsUseCase
	InsightsUC         *usecase.InsightsUseCase
}

func Setup(r *gin.Engine, deps Deps) {
	userH := handler.NewUserHandler(deps.UserUC)
	r.POST("/api/v1/register", userH.Register)
	r.POST("/api/v1/login", userH.Login)

	api := r.Group("/api/auth/v1", middleware.AuthRequired())

	api.GET("/users", userH.GetAll)
	api.GET("/users/me", userH.GetMe)
	api.PATCH("/users/profile", userH.UpdateProfile)

	txH := handler.NewTransactionHandler(deps.TransactionUC, deps.NotificationUC)
	api.GET("/transactions", txH.GetAll)
	api.POST("/transactions", txH.Create)
	api.GET("/transactions/monthly", txH.GetMonthlyExpenses)
	api.GET("/transactions/monthly-income", txH.GetMonthlyIncome)
	api.GET("/transactions/monthly-summary", txH.GetMonthlySummary)
	api.GET("/transactions/export", txH.Export)
	api.POST("/transactions/import", txH.Import)
	api.PUT("/transactions/:id", txH.Update)
	api.DELETE("/transactions/:id", txH.Delete)

	bH := handler.NewBudgetHandler(deps.BudgetUC)
	api.GET("/budgets", bH.GetAll)
	api.POST("/budgets", bH.Create)
	api.GET("/budgets/usage", bH.GetUsage)
	api.GET("/budgets/:id", bH.GetByID)
	api.PUT("/budgets/:id", bH.Update)
	api.DELETE("/budgets/:id", bH.Delete)

	gH := handler.NewGoalHandler(deps.GoalUC)
	api.GET("/goals", gH.GetAll)
	api.POST("/goals", gH.Create)
	api.GET("/goals/overview", gH.GetOverview)
	api.GET("/goals/milestones", gH.GetMilestones)
	api.GET("/goals/:id", gH.GetByID)
	api.PUT("/goals/:id", gH.Update)
	api.DELETE("/goals/:id", gH.Delete)
	api.PATCH("/goals/contribute", gH.Contribute)

	dH := handler.NewDashboardHandler(deps.DashboardUC)
	api.GET("/dashboard", dH.Get)

	cH := handler.NewChatHandler(deps.ChatUC)
	api.POST("/chat", cH.Ask)
	api.POST("/chat/stream", cH.AskStream)
	api.GET("/chat/history", cH.GetHistory)
	api.DELETE("/chat/history", cH.ClearHistory)

	mlH := handler.NewMLHandler(deps.MLUC)
	api.GET("/ml/analysis", mlH.GetAnalysis)
	api.GET("/ml/anomaly", mlH.GetAnomaly)
	api.GET("/ml/forecast", mlH.GetForecast)
	api.POST("/ml/forecast/start", mlH.StartForecast)
	api.GET("/ml/forecast/status/:job_id", mlH.GetForecastStatus)
	api.GET("/ml/insights", mlH.GetInsights)

	fpH := handler.NewFinancialProfileHandler(deps.FinancialProfileUC)
	api.POST("/financial-profile", fpH.Upsert)
	api.GET("/financial-profile", fpH.Get)

	nH := handler.NewNotificationHandler(deps.NotificationUC)
	api.GET("/notifications", nH.GetAll)
	api.PATCH("/notifications/read-all", nH.MarkAllRead)
	api.PATCH("/notifications/:id/read", nH.MarkRead)
	api.DELETE("/notifications/:id", nH.Delete)
	api.GET("/notifications/preferences", nH.GetPreferences)
	api.PUT("/notifications/preferences", nH.UpdatePreferences)
	// Frontend-compatible alias
	api.GET("/notification-settings", nH.GetPreferences)
	api.POST("/notification-settings", nH.UpdatePreferences)

	aH := handler.NewActivityLogHandler(deps.ActivityLogUC)
	api.GET("/activity", aH.GetActivity)

	rH := handler.NewReportsHandler(deps.ReportsUC)
	api.GET("/reports/monthly-summary", rH.GetMonthlySummary)
	api.GET("/reports/category-breakdown", rH.GetCategoryBreakdown)
	api.GET("/reports/savings-rate", rH.GetSavingsRate)
	api.GET("/reports/net-worth", rH.GetNetWorth)
	api.GET("/reports/month-comparison", rH.GetMonthComparison)
	// Year-based report endpoints (frontend chart format)
	api.GET("/reports/income-expense-trend", rH.GetIncomeExpenseTrend)
	api.GET("/reports/networth-history", rH.GetNetworthHistory)
	api.GET("/reports/savings-rate-history", rH.GetSavingsRateHistory)
	api.GET("/reports/month-comparison-v2", rH.GetMonthComparisonByDate)
	api.GET("/transactions/category-breakdown", rH.GetCategoryBreakdownDetailed)

	iH := handler.NewInsightsHandler(deps.InsightsUC)
	api.GET("/financial-health", iH.GetFinancialHealth)
	api.GET("/insights", iH.GetInsights)
	api.GET("/recommendations", iH.GetRecommendations)

	aiCoachH := handler.NewAICoachHandler(deps.InsightsUC, deps.FinancialProfileUC)
	api.GET("/ai-coach/context", aiCoachH.GetContext)
}
