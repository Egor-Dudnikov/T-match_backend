// Package models defines the data structures used across the T-match backend.
package models

import "time"

// AdminStats is the aggregate statistics returned by the admin panel.
type AdminStats struct {
	TotalInterns     int `json:"total_interns"`
	TotalCompanies   int `json:"total_companies"`
	TotalInternships int `json:"total_internships"`
	TotalResponses   int `json:"total_responses"`

	ResponsesPending   int `json:"responses_pending"`
	ResponsesReviewing int `json:"responses_reviewing"`
	ResponsesAccepted  int `json:"responses_accepted"`
	ResponsesRejected  int `json:"responses_rejected"`

	NewUsers7Days       int `json:"new_users_7_days"`
	NewInternships7Days int `json:"new_internships_7_days"`
	NewResponses7Days   int `json:"new_responses_7_days"`

	UsersOnline int `json:"users_online"`
}

// UserBan is a ban record applied to a user.
type UserBan struct {
	ID       int       `json:"id"`
	UserID   int       `json:"user_id"`
	Reason   string    `json:"reason"`
	BannedBy int       `json:"banned_by"`
	BannedAt time.Time `json:"banned_at"`
}

// AdminBanRequest is the payload for banning a user.
type AdminBanRequest struct {
	Reason string `json:"reason" validate:"required,min=1,max=500"`
}

// BanResponse is the ban result returned by the admin API.
type BanResponse struct {
	Reason string `json:"reason"`
}
