package main

import (
	"T-match_backend/internal/http"
	"T-match_backend/internal/rw"
	"encoding/json"
	"log"
	stdhttp "net/http"
	"os"

	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

func main() {
	configJson, err := os.Open("../configs/configuration.json")
	if err != nil {
		log.Fatalln("Config not found", err)
	}
	defer configJson.Close()

	config := rw.Config{}
	decoderJson := json.NewDecoder(configJson)
	err = decoderJson.Decode(&config)
	if err != nil {
		log.Fatalln(err)
	}

	db, err := rw.PingDatabase(config.DbConfig)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()
	app := &http.App{
		Db:  db,
		Log: log.Default(),
	}

	router := http.NewRouter(app)

	port := config.ServerConfig.Port
	addr := config.ServerConfig.Host
	log.Printf("Starting server at port %s, address %s", port, addr)
	if err := stdhttp.ListenAndServe(port, router); err != nil {
		log.Fatalln(err)
	}
}
