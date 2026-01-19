package http

import (
	"encoding/base32"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"firegoals/internal/auth"
	"firegoals/internal/repo"

	"github.com/go-chi/chi/v5"
)

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

type syncPushRequest struct {
	WorkspaceID string                      `json:"workspace_id"`
	Changes     map[string][]map[string]any `json:"changes"`
}

func (a *API) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password required")
		return
	}
	userID, err := a.Service.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, entityResponse{ID: userID})
}

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	accessToken, refreshToken, err := a.Service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	writeJSON(w, http.StatusOK, loginResponse{AccessToken: accessToken, RefreshToken: refreshToken})
}

func (a *API) handleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing user")
		return
	}
	id, email, err := a.Repo.GetUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "email": email})
}

func (a *API) handleListWorkspaces(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing user")
		return
	}
	ids, err := a.Repo.ListUserWorkspaces(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list workspaces")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"workspaces": ids})
}

func (a *API) handleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing user")
		return
	}
	var req workspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name required")
		return
	}
	workspaceType := req.Type
	if workspaceType == "" {
		workspaceType = "shared"
	}
	id, err := a.Repo.CreateWorkspace(r.Context(), req.Name, workspaceType, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create workspace")
		return
	}
	writeJSON(w, http.StatusCreated, entityResponse{ID: id})
}

func (a *API) handleWorkspaceBalance(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "id")
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing user")
		return
	}
	allowed, err := a.Repo.UserInWorkspace(r.Context(), userID, workspaceID)
	if err != nil || !allowed {
		writeError(w, http.StatusForbidden, "not allowed")
		return
	}
	balance, err := a.Repo.GetWorkspaceBalance(r.Context(), workspaceID)
	if err != nil {
		writeError(w, http.StatusNotFound, "balance not found")
		return
	}
	writeJSON(w, http.StatusOK, workspaceBalanceResponse{WorkspaceID: workspaceID, Balance: balance})
}

func (a *API) handleCreateInvite(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "id")
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing user")
		return
	}
	allowed, err := a.Repo.UserInWorkspace(r.Context(), userID, workspaceID)
	if err != nil || !allowed {
		writeError(w, http.StatusForbidden, "not allowed")
		return
	}
	code := randomCode()
	expiresAt := time.Now().Add(48 * time.Hour)
	if err := a.Repo.CreateInvite(r.Context(), workspaceID, userID, code, expiresAt); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create invite")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"code": code, "expires_at": expiresAt})
}

func (a *API) handleAcceptInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing user")
		return
	}
	var req inviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "code required")
		return
	}
	workspaceID, expiresAt, err := a.Repo.GetInvite(r.Context(), req.Code)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "invite not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to accept invite")
		return
	}
	if time.Now().After(expiresAt) {
		writeError(w, http.StatusBadRequest, "invite expired")
		return
	}
	if err := a.Repo.AddWorkspaceMember(r.Context(), workspaceID, userID, "member"); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to join workspace")
		return
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
		writeError(w, http.StatusInternalServerError, "failed to list goals")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"goals": goals})
}

func (a *API) handleCreateGoal(w http.ResponseWriter, r *http.Request) {
	var req goalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.Title == "" || req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "workspace_id and title required")
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
		writeError(w, http.StatusInternalServerError, "failed to create goal")
		return
	}
	writeJSON(w, http.StatusCreated, entityResponse{ID: id})
}

func (a *API) handleUpdateGoal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req goalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, req.WorkspaceID) {
		return
	}
	if err := a.Repo.UpdateGoal(r.Context(), id, req.WorkspaceID, req.Title, req.Description, req.Period, req.Status, req.StartDate, req.EndDate); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "goal not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update goal")
		return
	}
	writeJSON(w, http.StatusOK, entityResponse{ID: id})
}

func (a *API) handleDeleteGoal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	workspaceID := r.URL.Query().Get("workspace_id")
	if workspaceID == "" {
		writeError(w, http.StatusBadRequest, "workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, workspaceID) {
		return
	}
	if err := a.Repo.DeleteGoal(r.Context(), id, workspaceID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "goal not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete goal")
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
		writeError(w, http.StatusInternalServerError, "failed to list tasks")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
}

func (a *API) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req taskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.Title == "" || req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "workspace_id and title required")
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
		writeError(w, http.StatusInternalServerError, "failed to create task")
		return
	}
	writeJSON(w, http.StatusCreated, entityResponse{ID: id})
}

func (a *API) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req taskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, req.WorkspaceID) {
		return
	}
	if err := a.Repo.UpdateTask(r.Context(), id, req.WorkspaceID, req.GoalID, req.Title, req.Description, req.DueDate, req.RepeatRule, req.Value, req.Status); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update task")
		return
	}
	writeJSON(w, http.StatusOK, entityResponse{ID: id})
}

func (a *API) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	workspaceID := r.URL.Query().Get("workspace_id")
	if workspaceID == "" {
		writeError(w, http.StatusBadRequest, "workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, workspaceID) {
		return
	}
	if err := a.Repo.DeleteTask(r.Context(), id, workspaceID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete task")
		return
	}
	writeJSON(w, http.StatusOK, entityResponse{ID: id})
}

func (a *API) handleCompleteTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		WorkspaceID string `json:"workspace_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, req.WorkspaceID) {
		return
	}
	userID, _ := auth.UserIDFromContext(r.Context())
	value, err := a.Repo.CompleteTask(r.Context(), id, req.WorkspaceID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to complete task")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"earned": value})
}

func (a *API) handleListRewards(w http.ResponseWriter, r *http.Request) {
	workspaceID := r.URL.Query().Get("workspace_id")
	if !a.authorizeWorkspace(w, r, workspaceID) {
		return
	}
	rewards, err := a.Repo.ListRewards(r.Context(), workspaceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list rewards")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rewards": rewards})
}

func (a *API) handleCreateReward(w http.ResponseWriter, r *http.Request) {
	var req rewardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.Title == "" || req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "workspace_id and title required")
		return
	}
	if !a.authorizeWorkspace(w, r, req.WorkspaceID) {
		return
	}
	id, err := a.Repo.CreateReward(r.Context(), req.WorkspaceID, req.Title, req.Description, req.Cost, req.IsShared, req.CooldownHours)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create reward")
		return
	}
	writeJSON(w, http.StatusCreated, entityResponse{ID: id})
}

func (a *API) handleUpdateReward(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req rewardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, req.WorkspaceID) {
		return
	}
	if err := a.Repo.UpdateReward(r.Context(), id, req.WorkspaceID, req.Title, req.Description, req.Cost, req.IsShared, req.CooldownHours); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "reward not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update reward")
		return
	}
	writeJSON(w, http.StatusOK, entityResponse{ID: id})
}

func (a *API) handleDeleteReward(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	workspaceID := r.URL.Query().Get("workspace_id")
	if workspaceID == "" {
		writeError(w, http.StatusBadRequest, "workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, workspaceID) {
		return
	}
	if err := a.Repo.DeleteReward(r.Context(), id, workspaceID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "reward not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete reward")
		return
	}
	writeJSON(w, http.StatusOK, entityResponse{ID: id})
}

func (a *API) handleBuyReward(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		WorkspaceID string `json:"workspace_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, req.WorkspaceID) {
		return
	}
	userID, _ := auth.UserIDFromContext(r.Context())
	cost, err := a.Repo.BuyReward(r.Context(), id, req.WorkspaceID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "reward not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to buy reward")
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
		writeError(w, http.StatusInternalServerError, "failed to list achievements")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"achievements": achievements})
}

func (a *API) handleCreateAchievement(w http.ResponseWriter, r *http.Request) {
	var req achievementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.Title == "" || req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "workspace_id and title required")
		return
	}
	if !a.authorizeWorkspace(w, r, req.WorkspaceID) {
		return
	}
	id, err := a.Repo.CreateAchievement(r.Context(), req.WorkspaceID, req.Title, req.Description, req.ImageURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create achievement")
		return
	}
	writeJSON(w, http.StatusCreated, entityResponse{ID: id})
}

func (a *API) handleUpdateAchievement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req achievementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, req.WorkspaceID) {
		return
	}
	if err := a.Repo.UpdateAchievement(r.Context(), id, req.WorkspaceID, req.Title, req.Description, req.ImageURL, req.AchievedAt); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "achievement not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update achievement")
		return
	}
	writeJSON(w, http.StatusOK, entityResponse{ID: id})
}

func (a *API) handleDeleteAchievement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	workspaceID := r.URL.Query().Get("workspace_id")
	if workspaceID == "" {
		writeError(w, http.StatusBadRequest, "workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, workspaceID) {
		return
	}
	if err := a.Repo.DeleteAchievement(r.Context(), id, workspaceID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "achievement not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete achievement")
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
		writeError(w, http.StatusBadRequest, "since required")
		return
	}
	since, err := time.Parse(time.RFC3339, sinceStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid since")
		return
	}
	changes, err := a.Repo.ListSyncChanges(r.Context(), workspaceID, since)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to sync")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"changes": changes, "server_time": time.Now().UTC()})
}

func (a *API) handleSyncPush(w http.ResponseWriter, r *http.Request) {
	var req syncPushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "workspace_id required")
		return
	}
	if !a.authorizeWorkspace(w, r, req.WorkspaceID) {
		return
	}
	for table, entries := range req.Changes {
		for _, entry := range entries {
			entry["workspace_id"] = req.WorkspaceID
			if err := a.Repo.UpsertEntity(r.Context(), table, entry); err != nil {
				writeError(w, http.StatusBadRequest, "failed to upsert")
				return
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) authorizeWorkspace(w http.ResponseWriter, r *http.Request, workspaceID string) bool {
	if workspaceID == "" {
		writeError(w, http.StatusBadRequest, "workspace_id required")
		return false
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing user")
		return false
	}
	allowed, err := a.Repo.UserInWorkspace(r.Context(), userID, workspaceID)
	if err != nil || !allowed {
		writeError(w, http.StatusForbidden, "not allowed")
		return false
	}
	return true
}

func randomCode() string {
	return base32Encoder().EncodeToString([]byte(time.Now().Format("20060102150405.000")))
}

func base32Encoder() *base32.Encoding {
	return base32.StdEncoding.WithPadding(base32.NoPadding)
}
