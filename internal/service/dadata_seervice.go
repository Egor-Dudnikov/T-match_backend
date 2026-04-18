package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func ValidTIN(TIN string) (models.CompanyData, error) {
	company := models.CompanyData{}

	requestBody, err := json.Marshal(map[string]string{
		"query": TIN,
	})
	if err != nil {
		return company, fmt.Errorf("%w: %v", apierrors.ErrJSONDecodeFailed, err)
	}

	req, err := http.NewRequest("POST", "https://suggestions.dadata.ru/suggestions/api/4_1/rs/findById/party", bytes.NewBuffer(requestBody))
	if err != nil {
		return company, fmt.Errorf("%w: %v", apierrors.ErrInternalServer, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", os.Getenv("DA_DATA_API_KEY")))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return company, fmt.Errorf("%w: %v", apierrors.ErrBadGateway, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return company, fmt.Errorf("%w: %v", apierrors.ErrInternalServer, err)
	}

	if resp.StatusCode == http.StatusOK {
		var res map[string]interface{}
		err := json.Unmarshal(body, &res)
		if err != nil {
			return company, fmt.Errorf("%w: %v", apierrors.ErrJSONDecodeFailed, err)
		}

		suggestions := res["suggestions"].([]interface{})
		if len(suggestions) == 0 {
			return company, apierrors.ErrCompanyNotExists
		}

		first := suggestions[0].(map[string]interface{})
		data := first["data"].(map[string]interface{})

		inn := data["inn"].(string)
		kpp := data["kpp"].(string)
		ogrn := data["ogrn"].(string)
		okved := data["okved"].(string)
		branchType := data["branch_type"].(string)

		name := data["name"].(map[string]interface{})
		shortName := name["short_with_opf"].(string)

		state := data["state"].(map[string]interface{})
		status := state["status"].(string)

		management := data["management"].(map[string]interface{})
		director := management["name"].(string)
		directorPost := management["post"].(string)

		address := data["address"].(map[string]interface{})
		addrValue := address["value"].(string)

		company = models.CompanyData{
			Inn:          inn,
			Kpp:          kpp,
			Ogrn:         ogrn,
			Okved:        okved,
			BranchType:   branchType,
			ShortName:    shortName,
			Status:       status,
			Director:     director,
			DirectorPost: directorPost,
			Address:      addrValue,
		}

		if company.Status != "ACTIVE" {
			return company, apierrors.ErrCompanyNotExists
		}

		return company, nil

	}
	return company, apierrors.ErrCompanyNotExists
}
