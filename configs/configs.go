package configs

import (
	"T-match_backend/internal/rw"
	"encoding/json"
	"log"
	"os"
)

func PingConfig() (rw.Config, error) {
	configJson, err := os.Open("../configs/configuration.json")
	if err != nil {
		log.Fatalln("Config not found", err)
	}
	defer configJson.Close()

	config := rw.Config{}
	decoderJson := json.NewDecoder(configJson)
	err = decoderJson.Decode(&config)
	if err != nil {
		return config, err
	}
	return config, nil
}
