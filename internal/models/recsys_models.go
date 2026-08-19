// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package models

// RecsysGeo is the geo coordinates sent to the recommendation service.
type RecsysGeo struct {
	GeoLen float64 `json:"geo_len"`
	GeoLot float64 `json:"geo_lot"`
}

// NewInternshipRec is the internship payload for the recommendation service.
type NewInternshipRec struct {
	ID     int     `json:"id"`
	GeoLen float64 `json:"geo_len"`
	GeoLot float64 `json:"geo_lot"`
}

// RecsysSkill is a skill payload for the recommendation service.
type RecsysSkill struct {
	SkillID int `json:"skill_id"`
}

// RecsysAction is a user-internship interaction reported to the recommendation service.
type RecsysAction struct {
	UserID       int    `json:"user_id"`
	InternshipID int    `json:"internship_id"`
	ActionType   string `json:"action_type"`
}

// Recommendation is an internship recommended to a user.
type Recommendation struct {
	InternshipID    int      `json:"internship_id"`
	Score           float64  `json:"score"`
	GeoSimilarity   float64  `json:"geo_similarity"`
	SkillSimilarity float64  `json:"skill_similarity"`
	DistanceKm      float64  `json:"distance_km"`
	AlsScore        *float64 `json:"als_score"`
}
