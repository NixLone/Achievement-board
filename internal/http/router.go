package http

import (
	"net/http"
	"time"

	"firegoals/internal/auth"
	"firegoals/internal/repo"
	"firegoals/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type API struct {
	Repo    *repo.Repo
	Service *service.Service
	Auth    *auth.Manager
	Origins []string
}

func (a *API) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(loggingMiddleware)
	r.Use(a.corsMiddleware)

	r.Get("/health", a.handleHealth)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", a.handleRegister)
		r.Post("/login", a.handleLogin)
	})

	r.Group(func(r chi.Router) {
		r.Use(a.authMiddleware)
		r.Get("/me", a.handleMe)
		r.Get("/settings", a.handleGetSettings)
		r.Put("/settings", a.handleUpdateSettings)
		r.Get("/workspaces", a.handleListWorkspaces)
		r.Post("/workspaces", a.handleCreateWorkspace)
		r.Get("/workspaces/{id}/balance", a.handleWorkspaceBalance)
		r.Get("/workspaces/{id}/members", a.handleListWorkspaceMembers)
		r.Post("/workspaces/{id}/invite", a.handleCreateInvite)
		r.Post("/invites/accept", a.handleAcceptInvite)

		r.Route("/goals", func(r chi.Router) {
			r.Get("/", a.handleListGoals)
			r.Post("/", a.handleCreateGoal)
			r.Put("/{id}", a.handleUpdateGoal)
			r.Delete("/{id}", a.handleDeleteGoal)
		})
		r.Route("/tasks", func(r chi.Router) {
			r.Get("/", a.handleListTasks)
			r.Post("/", a.handleCreateTask)
			r.Put("/{id}", a.handleUpdateTask)
			r.Delete("/{id}", a.handleDeleteTask)
			r.Post("/{id}/complete", a.handleCompleteTask)
		})
		r.Route("/rewards", func(r chi.Router) {
			r.Get("/", a.handleListRewards)
			r.Get("/purchases", a.handleListRewardPurchases)
			r.Post("/", a.handleCreateReward)
			r.Put("/{id}", a.handleUpdateReward)
			r.Delete("/{id}", a.handleDeleteReward)
			r.Post("/{id}/buy", a.handleBuyReward)
		})
		r.Route("/achievements", func(r chi.Router) {
			r.Get("/", a.handleListAchievements)
			r.Post("/", a.handleCreateAchievement)
			r.Put("/{id}", a.handleUpdateAchievement)
			r.Delete("/{id}", a.handleDeleteAchievement)
		})
		r.Get("/sync", a.handleSyncPull)
		r.Post("/sync", a.handleSyncPush)
	})

	return r
}
