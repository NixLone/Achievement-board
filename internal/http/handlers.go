package http

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"firegoals/internal/auth"
	"firegoals/internal/repo"

	"github.com/go-chi/chi/v5"
)

const maxBodyBytes = 1 << 20

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type workspaceRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type inviteRequest struct {
	Code string `json:"code"`
}

type entityResponse struct {
	ID string `json:"id"`
}

type goalRequest struct {
	WorkspaceID string     `json:"workspace_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Period      string     `json:"period"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Status      string     `json:"status"`
}

type taskRequest struct {
	WorkspaceID string     `json:"workspace_id"`
	GoalID      *string    `json:"goal_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	RepeatRule  *string    `json:"repeat_rule"`
	Value       float64    `json:"value"`
	Status      string     `json:"status"`
}

type rewardRequest struct {
	WorkspaceID   string  `json:"workspace_id"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Cost          float64 `json:"cost"`
	IsShared      bool    `json:"is_shared"`
	CooldownHours *int    `json:"cooldown_hours"`
}

type achievementRequest struct {
	WorkspaceID string     `json:"workspace_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	ImageURL    *string    `json:"image_url"`
	AchievedAt  *time.Time `json:"achieved_at"`
}

type workspaceBalanceResponse struct {
	WorkspaceID string  `json:"workspace_id"`
	Balance     float64 `json:"balance"`
}

func (a *API) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Email and password required")
		return
	}
	userID, err := a.Service.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, "REGISTRATION_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, entityResponse{ID: userID})
}

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	accessToken, refreshToken, err := a.Service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid credentials")
		return
	}
	writeJSON(w, http.StatusOK, loginResponse{AccessToken: accessToken, RefreshToken: refreshToken})
}

func (a *API) handleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user")
		return
	}
	id, email, err := a.Repo.GetUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "email": email})
}

func (a *API) handleListWorkspaces(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user")
		return
	}
	ids, err := a.Repo.ListUserWorkspaces(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list workspaces")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"workspaces": ids})
}

func (a *API) handleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user")
		return
	}
	var req workspaceRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Name required")
		return
	}
	workspaceType := req.Type
	if workspaceType == "" {
		workspaceType = "shared"
	}
	id, err := a.Repo.CreateWorkspace(r.Context(), req.Name, workspaceType, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create workspace")
		return
	}
	writeJSON(w, http.StatusCreated, entityResponse{ID: id})
}

func (a *API) handleWorkspaceBalance(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "id")
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user")
		return
	}
	allowed, err := a.Repo.UserInWorkspace(r.Context(), userID, workspaceID)
	if err != nil || !allowed {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Not allowed")
		return
	}
	balance, err := a.Repo.GetWorkspaceBalance(r.Context(), workspaceID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Balance not found")
		return
	}
	writeJSON(w, http.StatusOK, workspaceBalanceResponse{WorkspaceID: workspaceID, Balance: balance})
}

func (a *API) handleCreateInvite(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "id")
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user")
		return
	}
	allowed, err := a.Repo.UserInWorkspace(r.Context(), userID, workspaceID)
	if err != nil || !allowed {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Not allowed")
		return
	}
	role, err := a.Repo.GetWorkspaceRole(r.Context(), userID, workspaceID)
	if err != nil || role != "owner" {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Only owner can invite")
		return
	}
	code, err := randomCode()
	if err != nil {
		log.Printf("invite code generation failed: %v", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate invite")
		return
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := a.Repo.CreateInvite(r.Context(), workspaceID, userID, code, expiresAt); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create invite")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"code": code, "expires_at": expiresAt})
}

func (a *API) handleAcceptInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user")
		return
	}
	var req inviteRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Code required")
		return
	}
	workspaceID, err := a.Repo.AcceptInvite(r.Context(), req.Code, userID)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrInviteExpired):
			writeError(w, http.StatusBadRequest, "INVITE_EXPIRED", "Invite expired")
			return
		case errors.Is(err, repo.ErrInviteUsed):
			writeError(w, http.StatusBadRequest, "INVITE_USED", "Invite already used")
			return
		case errors.Is(err, repo.ErrNotFound):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Invite not found")
			return
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to accept invite")
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"workspace_id": workspaceID})
}

func (a *API) handleListGoals(w http.ResponseWriter, r *http.Request) {
	workspaceID := r.URL.Query().Get("workspace_id")
	if !a.authorizeWorkspace(w, r, workspaceID) {
		return
	}
	goals, err := a.Repo.ListGoals(r.Context(), workspaceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list goals")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"goals": goals})
}

func (a *API) handleCreateGoal(w http.ResponseWriter, r *http.Request) {
	var req goalRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Title == "" || req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Workspace_id and title required")
		return
	}
	if !a.authorizeWorkspace(w, r, req.WorkspaceID) {
		return
	}
	status := req.Status
	if status == "" {
		status = "active"
	}
	id, err := a.Repo.CreateGoal(r.Context(), req.WorkspaceID, req.Title, req.Description, req.Period, status, req.StartDate, req.EndDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create goal")
		return
	}
	writeJSON(w, http.StatusCreated, entityResponse{ID: id})
}

func (a *API) handleUpdateGoal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req goalRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, req.WorkspaceID) {
		return
	}
	if err := a.Repo.UpdateGoal(r.Context(), id, req.WorkspaceID, req.Title, req.Description, req.Period, req.Status, req.StartDate, req.EndDate); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Goal not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update goal")
		return
	}
	writeJSON(w, http.StatusOK, entityResponse{ID: id})
}

func (a *API) handleDeleteGoal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	workspaceID := r.URL.Query().Get("workspace_id")
	if workspaceID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, workspaceID) {
		return
	}
	if err := a.Repo.DeleteGoal(r.Context(), id, workspaceID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Goal not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete goal")
		return
	}
	writeJSON(w, http.StatusOK, entityResponse{ID: id})
}

func (a *API) handleListTasks(w http.ResponseWriter, r *http.Request) {
	workspaceID := r.URL.Query().Get("workspace_id")
	if !a.authorizeWorkspace(w, r, workspaceID) {
		return
	}
	tasks, err := a.Repo.ListTasks(r.Context(), workspaceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list tasks")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
}

func (a *API) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req taskRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Title == "" || req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Workspace_id and title required")
		return
	}
	if !a.authorizeWorkspace(w, r, req.WorkspaceID) {
		return
	}
	status := req.Status
	if status == "" {
		status = "open"
	}
	id, err := a.Repo.CreateTask(r.Context(), req.WorkspaceID, req.GoalID, req.Title, req.Description, req.DueDate, req.RepeatRule, req.Value, status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create task")
		return
	}
	writeJSON(w, http.StatusCreated, entityResponse{ID: id})
}

func (a *API) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req taskRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, req.WorkspaceID) {
		return
	}
	if err := a.Repo.UpdateTask(r.Context(), id, req.WorkspaceID, req.GoalID, req.Title, req.Description, req.DueDate, req.RepeatRule, req.Value, req.Status); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update task")
		return
	}
	writeJSON(w, http.StatusOK, entityResponse{ID: id})
}

func (a *API) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	workspaceID := r.URL.Query().Get("workspace_id")
	if workspaceID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, workspaceID) {
		return
	}
	if err := a.Repo.DeleteTask(r.Context(), id, workspaceID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete task")
		return
	}
	writeJSON(w, http.StatusOK, entityResponse{ID: id})
}

func (a *API) handleCompleteTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		WorkspaceID string `json:"workspace_id"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, req.WorkspaceID) {
		return
	}
	userID, _ := auth.UserIDFromContext(r.Context())
	value, completed, err := a.Repo.CompleteTask(r.Context(), id, req.WorkspaceID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to complete task")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"earned": value, "completed": completed})
}

func (a *API) handleListRewards(w http.ResponseWriter, r *http.Request) {
	workspaceID := r.URL.Query().Get("workspace_id")
	if !a.authorizeWorkspace(w, r, workspaceID) {
		return
	}
	rewards, err := a.Repo.ListRewards(r.Context(), workspaceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list rewards")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rewards": rewards})
}

func (a *API) handleCreateReward(w http.ResponseWriter, r *http.Request) {
	var req rewardRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Title == "" || req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Workspace_id and title required")
		return
	}
	if !a.authorizeWorkspace(w, r, req.WorkspaceID) {
		return
	}
	id, err := a.Repo.CreateReward(r.Context(), req.WorkspaceID, req.Title, req.Description, req.Cost, req.IsShared, req.CooldownHours)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create reward")
		return
	}
	writeJSON(w, http.StatusCreated, entityResponse{ID: id})
}

func (a *API) handleUpdateReward(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req rewardRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, req.WorkspaceID) {
		return
	}
	if err := a.Repo.UpdateReward(r.Context(), id, req.WorkspaceID, req.Title, req.Description, req.Cost, req.IsShared, req.CooldownHours); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Reward not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update reward")
		return
	}
	writeJSON(w, http.StatusOK, entityResponse{ID: id})
}

func (a *API) handleDeleteReward(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	workspaceID := r.URL.Query().Get("workspace_id")
	if workspaceID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, workspaceID) {
		return
	}
	if err := a.Repo.DeleteReward(r.Context(), id, workspaceID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Reward not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete reward")
		return
	}
	writeJSON(w, http.StatusOK, entityResponse{ID: id})
}

func (a *API) handleBuyReward(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		WorkspaceID string `json:"workspace_id"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, req.WorkspaceID) {
		return
	}
	userID, _ := auth.UserIDFromContext(r.Context())
	cost, err := a.Repo.BuyReward(r.Context(), id, req.WorkspaceID, userID)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrInsufficientFunds):
			writeError(w, http.StatusBadRequest, "INSUFFICIENT_FUNDS", "Недостаточно огоньков")
			return
		case errors.Is(err, repo.ErrNotFound):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Reward not found")
			return
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to buy reward")
			return
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"spent": cost})
}

func (a *API) handleListAchievements(w http.ResponseWriter, r *http.Request) {
	workspaceID := r.URL.Query().Get("workspace_id")
	if !a.authorizeWorkspace(w, r, workspaceID) {
		return
	}
	achievements, err := a.Repo.ListAchievements(r.Context(), workspaceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list achievements")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"achievements": achievements})
}

func (a *API) handleCreateAchievement(w http.ResponseWriter, r *http.Request) {
	var req achievementRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Title == "" || req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Workspace_id and title required")
		return
	}
	if !a.authorizeWorkspace(w, r, req.WorkspaceID) {
		return
	}
	id, err := a.Repo.CreateAchievement(r.Context(), req.WorkspaceID, req.Title, req.Description, req.ImageURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create achievement")
		return
	}
	writeJSON(w, http.StatusCreated, entityResponse{ID: id})
}

func (a *API) handleUpdateAchievement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req achievementRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, req.WorkspaceID) {
		return
	}
	if err := a.Repo.UpdateAchievement(r.Context(), id, req.WorkspaceID, req.Title, req.Description, req.ImageURL, req.AchievedAt); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Achievement not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update achievement")
		return
	}
	writeJSON(w, http.StatusOK, entityResponse{ID: id})
}

func (a *API) handleDeleteAchievement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	workspaceID := r.URL.Query().Get("workspace_id")
	if workspaceID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, workspaceID) {
		return
	}
	if err := a.Repo.DeleteAchievement(r.Context(), id, workspaceID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Achievement not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete achievement")
		return
	}
	writeJSON(w, http.StatusOK, entityResponse{ID: id})
}

func (a *API) handleSyncPull(w http.ResponseWriter, r *http.Request) {
	workspaceID := r.URL.Query().Get("workspace_id")
	if !a.authorizeWorkspace(w, r, workspaceID) {
		return
	}
	sinceStr := r.URL.Query().Get("since")
	if sinceStr == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Since required")
		return
	}
	since, err := time.Parse(time.RFC3339, sinceStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid since")
		return
	}
	cursorTime := time.Now().UTC()
	changes, err := a.Repo.ListSyncChanges(r.Context(), workspaceID, since, cursorTime)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to sync")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"changes": changes, "server_time": cursorTime})
}

func (a *API) handleSyncPush(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusBadRequest, "SYNC_PUSH_DISABLED", "Sync push is disabled in MVP v2")
}

func (a *API) authorizeWorkspace(w http.ResponseWriter, r *http.Request, workspaceID string) bool {
	if workspaceID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Workspace_id required")
		return false
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user")
		return false
	}
	allowed, err := a.Repo.UserInWorkspace(r.Context(), userID, workspaceID)
	if err != nil || !allowed {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Not allowed")
		return false
	}
	return true
}

func randomCode() (string, error) {
	data, err := randomBytes(20)
	if err != nil {
		return "", err
	}
	return base32Encoder().EncodeToString(data), nil
}

func base32Encoder() *base32.Encoding {
	return base32.StdEncoding.WithPadding(base32.NoPadding)
}

func randomBytes(length int) ([]byte, error) {
	buf := make([]byte, length)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid payload")
		return false
	}
	return true
}
