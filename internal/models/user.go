package models

import "time"

// User represents a user inside PostgreSQL database (Google OAuth & Subscription metadata)
type User struct {
	ID               string    `json:"id" db:"id"`
	Email            string    `json:"email" db:"email"`
	GoogleID         string    `json:"google_id" db:"google_id"`
	AccessToken      string    `json:"access_token" db:"access_token"`
	RefreshToken     string    `json:"refresh_token" db:"refresh_token"`
	TokenExpiry      time.Time `json:"token_expiry" db:"token_expiry"`
	SubscriptionType string    `json:"subscription_type" db:"subscription_type"` // 'free' or 'premium'
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// UsageLog represents a single API consumption record for quota tracking
type UsageLog struct {
	ID         int64     `json:"id" db:"id"`
	UserID     string    `json:"user_id" db:"user_id"`
	ActionType string    `json:"action_type" db:"action_type"`
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
}
