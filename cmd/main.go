package main

import (
	"T-match_backend/configs"
	"T-match_backend/internal/cash"
	"T-match_backend/internal/handlers"
	"T-match_backend/internal/repository"
	"log"
	"net/http"

	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

func main() {
	config, err := configs.PingConfig()
	if err != nil {
		log.Fatalln(err)
	}

	db, err := repository.PingDatabase(config.DbConfig)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	dbr, err := cash.PingRedis(config.RedisConfig)
	if err != nil {
		log.Fatalln(err)
	}
	defer dbr.Close()

	app := &handlers.App{
		Db:  db,
		Dbr: dbr,
		Cfg: config,
	}

	router := handlers.NewRouter(app)

	port := config.ServerConfig.Port
	addr := config.ServerConfig.Host
	log.Printf("Starting server at port %s, address %s", port, addr)
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalln(err)
	}
}
