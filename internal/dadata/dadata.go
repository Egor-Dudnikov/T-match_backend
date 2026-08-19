// Package dadata provides a client for the DaData company lookup API.
package dadata

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/tidwall/gjson"
)

// Client is a client for the DaData API.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a DaData client using the API key from the environment.
func NewClient() *Client {
	return &Client{
		apiKey: os.Getenv("DA_DATA_API_KEY"),
		httpClient: &http.Client{
			Timeout: constants.DadataTimeout,
		},
	}
}

// MakeRequest looks up the company with the given TIN and returns the raw response body and status code.
func (c *Client) MakeRequest(TIN string) ([]byte, int, error) {
	requestBody, err := json.Marshal(map[string]string{
		"query": TIN,
	})
	if err != nil {
		return []byte{}, 500, apierrors.Wrap(apierrors.ErrJSONDecodeFailed, err)
	}

	req, err := http.NewRequest("POST", "https://suggestions.dadata.ru/suggestions/api/4_1/rs/findById/party", bytes.NewBuffer(requestBody))
	if err != nil {
		return []byte{}, 500, apierrors.Wrap(apierrors.ErrInternalServer, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return []byte{}, resp.StatusCode, apierrors.Wrap(apierrors.ErrBadGateway, err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("dadata: close response body: %v", cerr)
		}
	}()

	respJSON, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, resp.StatusCode, apierrors.Wrap(apierrors.ErrInternalServer, err)
	}
	return respJSON, resp.StatusCode, nil
}

// ValidTIN validates the given TIN and returns the matching active company data.
func (c *Client) ValidTIN(TIN string) (models.CompanyData, error) {
	company := models.CompanyData{}

	body, statusCode, err := c.MakeRequest(TIN)
	if err != nil {
		return company, err
	}

	if statusCode != http.StatusOK {
		return company, apierrors.ErrCompanyNotExists
	}

	suggestions := gjson.GetBytes(body, "suggestions")
	if !suggestions.Exists() || len(suggestions.Array()) == 0 {
		return company, apierrors.ErrCompanyNotExists
	}

	data := gjson.GetBytes(body, "suggestions.0.data")
	status := data.Get("state.status").String()
	if status != "ACTIVE" {
		return models.CompanyData{}, apierrors.ErrCompanyNotExists
	}

	company = models.CompanyData{
		Inn:        data.Get("inn").String(),
		Kpp:        data.Get("kpp").String(),
		Ogrn:       data.Get("ogrn").String(),
		Okved:      data.Get("okved").String(),
		BranchType: data.Get("branch_type").String(),
		ShortName:  data.Get("name.short_with_opf").String(),
		Status:     status,
		Address:    data.Get("address.value").String(),
	}

	if company.Inn != TIN {
		return models.CompanyData{}, apierrors.ErrCompanyNotExists
	}

	return company, nil
}
