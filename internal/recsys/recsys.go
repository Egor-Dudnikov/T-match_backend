// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

// Package recsys provides a client for the recommendation service API.
package recsys

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

// Client is a client for the recommendation service.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new recommendation service client with the given base URL.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: constants.RecsysTimeout},
	}
}

func (c *Client) do(ctx context.Context, method, path string, payload interface{}, out interface{}) error {
	if c == nil || c.baseURL == "" {
		return nil
	}

	var body io.Reader
	if payload != nil {
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return apierrors.Wrap(apierrors.ErrJSONEncodeFailed, err)
		}
		body = bytes.NewReader(payloadJSON)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrInternalServer, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrBadGateway, err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("recsys: close response body: %v", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return apierrors.Wrap(apierrors.ErrBadGateway, fmt.Errorf("recsys responded %d for %s %s", resp.StatusCode, method, path))
	}

	if out == nil {
		return nil
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrInternalServer, err)
	}
	if len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return apierrors.Wrap(apierrors.ErrJSONDecodeFailed, err)
	}
	return nil
}

// CreateUser registers a new user in the recommendation service.
func (c *Client) CreateUser(ctx context.Context, userID int) error {
	return c.do(ctx, http.MethodPost, "/intern", map[string]int{"id": userID}, nil)
}

// DeleteUser removes a user from the recommendation service.
func (c *Client) DeleteUser(ctx context.Context, userID int) error {
	return c.do(ctx, http.MethodDelete, "/intern/"+strconv.Itoa(userID), nil, nil)
}

// UpdateUserGeo updates the user's geographic coordinates in the recommendation service.
func (c *Client) UpdateUserGeo(ctx context.Context, userID int, geoLen, geoLot float64) error {
	path := "/intern/" + strconv.Itoa(userID)
	return c.do(ctx, http.MethodPut, path, models.RecsysGeo{GeoLen: geoLen, GeoLot: geoLot}, nil)
}

// CreateInternship registers a new internship in the recommendation service.
func (c *Client) CreateInternship(ctx context.Context, internshipID int, geoLen, geoLot float64) error {
	return c.do(ctx, http.MethodPost, "/internship", models.NewInternshipRec{
		ID:     internshipID,
		GeoLen: geoLen,
		GeoLot: geoLot,
	}, nil)
}

// DeleteInternship removes an internship from the recommendation service.
func (c *Client) DeleteInternship(ctx context.Context, internshipID int) error {
	return c.do(ctx, http.MethodDelete, "/internship/"+strconv.Itoa(internshipID), nil, nil)
}

// UpdateInternshipGeo updates the internship's geographic coordinates in the recommendation service.
func (c *Client) UpdateInternshipGeo(ctx context.Context, internshipID int, geoLen, geoLot float64) error {
	path := "/internship/" + strconv.Itoa(internshipID)
	return c.do(ctx, http.MethodPut, path, models.RecsysGeo{GeoLen: geoLen, GeoLot: geoLot}, nil)
}

// AddUserSkill adds a skill to a user in the recommendation service.
func (c *Client) AddUserSkill(ctx context.Context, userID, skillID int) error {
	path := "/intern/" + strconv.Itoa(userID) + "/skills"
	return c.do(ctx, http.MethodPost, path, models.RecsysSkill{SkillID: skillID}, nil)
}

// DeleteUserSkill removes a skill from a user in the recommendation service.
func (c *Client) DeleteUserSkill(ctx context.Context, userID, skillID int) error {
	path := "/intern/" + strconv.Itoa(userID) + "/skills"
	return c.do(ctx, http.MethodDelete, path, models.RecsysSkill{SkillID: skillID}, nil)
}

// AddInternshipSkill adds a skill to an internship in the recommendation service.
func (c *Client) AddInternshipSkill(ctx context.Context, internshipID, skillID int) error {
	path := "/internship/" + strconv.Itoa(internshipID) + "/skills"
	return c.do(ctx, http.MethodPost, path, models.RecsysSkill{SkillID: skillID}, nil)
}

// DeleteInternshipSkill removes a skill from an internship in the recommendation service.
func (c *Client) DeleteInternshipSkill(ctx context.Context, internshipID, skillID int) error {
	path := "/internship/" + strconv.Itoa(internshipID) + "/skills"
	return c.do(ctx, http.MethodDelete, path, models.RecsysSkill{SkillID: skillID}, nil)
}

// AddAction sends a user action for an internship to the recommendation service.
func (c *Client) AddAction(ctx context.Context, userID, internshipID int, actionType string) error {
	return c.do(ctx, http.MethodPost, "/actions", models.RecsysAction{
		UserID:       userID,
		InternshipID: internshipID,
		ActionType:   actionType,
	}, nil)
}

// GetRecommendations returns internship recommendations for the given user.
func (c *Client) GetRecommendations(ctx context.Context, userID int) ([]models.Recommendation, error) {
	recs := []models.Recommendation{}
	err := c.do(ctx, http.MethodGet, "/intern/"+strconv.Itoa(userID)+"/recommend", nil, &recs)
	if err != nil {
		return nil, err
	}
	return recs, nil
}
