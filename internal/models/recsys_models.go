// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package models

type RecsysGeo struct {
	GeoLen float64 `json:"geo_len"`
	GeoLot float64 `json:"geo_lot"`
}

type NewInternshipRec struct {
	ID     int     `json:"id"`
	GeoLen float64 `json:"geo_len"`
	GeoLot float64 `json:"geo_lot"`
}

type RecsysSkill struct {
	SkillID int `json:"skill_id"`
}

type RecsysAction struct {
	UserID       int    `json:"user_id"`
	InternshipID int    `json:"internship_id"`
	ActionType   string `json:"action_type"`
}

type Recommendation struct {
	InternshipID    int      `json:"internship_id"`
	Score           float64  `json:"score"`
	GeoSimilarity   float64  `json:"geo_similarity"`
	SkillSimilarity float64  `json:"skill_similarity"`
	DistanceKm      float64  `json:"distance_km"`
	AlsScore        *float64 `json:"als_score"`
}
