package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WorkspaceMember struct {
	WorkspaceID string    `json:"workspace_id"`
	UserID      string    `json:"user_id"`
	Role        string    `json:"role"`
	Permissions string    `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
}

type Goal struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspace_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Period      string     `json:"period"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
	Version     int        `json:"version"`
}

type Task struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspace_id"`
	GoalID      *string    `json:"goal_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	IsRecurring bool       `json:"is_recurring"`
	Weekdays    []int      `json:"recurrence_weekdays"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Timezone    *string    `json:"timezone"`
	RepeatRule  *string    `json:"repeat_rule"`
	Value       float64    `json:"value"`
	Status      string     `json:"status"`
	DoneAt      *time.Time `json:"done_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
	Version     int        `json:"version"`
}

type Reward struct {
	ID            string     `json:"id"`
	WorkspaceID   string     `json:"workspace_id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Cost          float64    `json:"cost"`
	IsShared      bool       `json:"is_shared"`
	CooldownHours *int       `json:"cooldown_hours"`
	OneTime       bool       `json:"one_time"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at"`
	Version       int        `json:"version"`
}

type RewardPurchase struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	RewardID    string    `json:"reward_id"`
	UserID      string    `json:"user_id"`
	Cost        float64   `json:"cost"`
	PurchasedAt time.Time `json:"purchased_at"`
	Note        *string   `json:"note"`
}

type Achievement struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspace_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	ImageURL    *string    `json:"image_url"`
	AchievedAt  *time.Time `json:"achieved_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
	Version     int        `json:"version"`
}

type Transaction struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	UserID      *string   `json:"user_id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Reason      string    `json:"reason"`
	EntityType  *string   `json:"entity_type"`
	EntityID    *string   `json:"entity_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type WorkspaceBalance struct {
	WorkspaceID string    `json:"workspace_id"`
	Balance     float64   `json:"balance"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type UserSettings struct {
	UserID              string    `json:"user_id"`
	Theme               string    `json:"theme"`
	LastActiveWorkspace *string   `json:"last_active_workspace"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type Invite struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspace_id"`
	Code        string     `json:"code"`
	CreatedBy   string     `json:"created_by_user_id"`
	ExpiresAt   time.Time  `json:"expires_at"`
	UsedAt      *time.Time `json:"used_at"`
	CreatedAt   time.Time  `json:"created_at"`
}
