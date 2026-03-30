package configs

import (
	"T-match_backend/internal/models"
	"encoding/json"
	"log"
	"os"
)

func PingConfig() (models.Config, error) {
	configJson, err := os.Open(os.Getenv("CONFIG_PATH"))
	if err != nil {
		log.Fatalln("Config not found", err)
	}
	defer configJson.Close()

	config := models.Config{}
	decoderJson := json.NewDecoder(configJson)
	err = decoderJson.Decode(&config)
	if err != nil {
		return config, err
	}
	return config, nil
}
