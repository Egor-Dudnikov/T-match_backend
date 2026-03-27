package main

import (
	"T-match_backend/configs"
	"T-match_backend/internal/cache"
	"T-match_backend/internal/handlers"
	"T-match_backend/internal/repository"
	"T-match_backend/internal/service"
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

	dbr, err := cache.PingRedis(config.RedisConfig)
	if err != nil {
		log.Fatalln(err)
	}
	defer dbr.Close()

	repo := repository.NewRepository(db)
	redis := cache.NewRedis(dbr)
	email := service.NewEmailClient(config.EmailConfig)

	app := service.NewAuthService(repo, redis, email)
	authHandler := handlers.NewAuthServiceHandler(app)

	router := handlers.NewRouter(authHandler)

	port := config.ServerConfig.Port
	addr := config.ServerConfig.Host
	log.Printf("Starting server at port %s, address %s", port, addr)
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalln(err)
	}
}
