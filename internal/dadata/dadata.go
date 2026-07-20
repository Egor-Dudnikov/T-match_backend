package dadata

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		apiKey: os.Getenv("DA_DATA_API_KEY"),
		httpClient: &http.Client{
			Timeout: constants.DadataTimeout,
		},
	}
}

func (c *Client) ValidTIN(TIN string) (models.CompanyData, error) {
	company := models.CompanyData{}

	requestBody, err := json.Marshal(map[string]string{
		"query": TIN,
	})
	if err != nil {
		return company, apierrors.Wrap(apierrors.ErrJSONDecodeFailed, err)
	}

	req, err := http.NewRequest("POST", "https://suggestions.dadata.ru/suggestions/api/4_1/rs/findById/party", bytes.NewBuffer(requestBody))
	if err != nil {
		return company, apierrors.Wrap(apierrors.ErrInternalServer, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return company, apierrors.Wrap(apierrors.ErrBadGateway, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return company, apierrors.Wrap(apierrors.ErrInternalServer, err)
	}

	if resp.StatusCode == http.StatusOK {
		var res map[string]interface{}
		err := json.Unmarshal(body, &res)
		if err != nil {
			return company, apierrors.Wrap(apierrors.ErrJSONDecodeFailed, err)
		}

		suggestionsRaw, ok := res["suggestions"]
		if !ok || suggestionsRaw == nil {
			return company, apierrors.ErrCompanyNotExists
		}

		suggestions, ok := suggestionsRaw.([]interface{})
		if !ok {
			return company, apierrors.ErrCompanyNotExists
		}

		first, ok := suggestions[0].(map[string]interface{})
		if !ok {
			return company, apierrors.ErrCompanyNotExists
		}

		data, ok := first["data"].(map[string]interface{})
		if !ok {
			return company, apierrors.ErrCompanyNotExists
		}

		inn, ok := data["inn"].(string)
		if !ok {
			return company, apierrors.ErrCompanyNotExists
		}
		kpp, ok := data["kpp"].(string)
		if !ok {
			return company, apierrors.ErrCompanyNotExists
		}
		ogrn, ok := data["ogrn"].(string)
		if !ok {
			return company, apierrors.ErrCompanyNotExists
		}
		okved, ok := data["okved"].(string)
		if !ok {
			return company, apierrors.ErrCompanyNotExists
		}
		branchType, ok := data["branch_type"].(string)
		if !ok {
			return company, apierrors.ErrCompanyNotExists
		}

		name, ok := data["name"].(map[string]interface{})
		if !ok {
			return company, apierrors.ErrCompanyNotExists
		}
		shortName, ok := name["short_with_opf"].(string)
		if !ok {
			return company, apierrors.ErrCompanyNotExists
		}

		state, ok := data["state"].(map[string]interface{})
		if !ok {
			return company, apierrors.ErrCompanyNotExists
		}
		status, ok := state["status"].(string)
		if !ok {
			return company, apierrors.ErrCompanyNotExists
		}

		address, ok := data["address"].(map[string]interface{})
		if !ok {
			return company, apierrors.ErrCompanyNotExists
		}
		addrValue, ok := address["value"].(string)
		if !ok {
			return company, apierrors.ErrCompanyNotExists
		}

		company = models.CompanyData{
			Inn:        inn,
			Kpp:        kpp,
			Ogrn:       ogrn,
			Okved:      okved,
			BranchType: branchType,
			ShortName:  shortName,
			Status:     status,
			Address:    addrValue,
		}

		if company.Status != "ACTIVE" {
			return company, apierrors.ErrCompanyNotExists
		}

		return company, nil

	}
	return company, apierrors.ErrCompanyNotExists
}
