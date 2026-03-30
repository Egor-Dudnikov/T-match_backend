package main

import (
	"T-match_backend/configs"
	"T-match_backend/internal/cache"
	"T-match_backend/internal/handlers"
	"T-match_backend/internal/repository"
	"T-match_backend/internal/service"
	"T-match_backend/internal/utils"
	"log"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

func main() {
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatalln("not JWT_SECRET in env")
	}
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
	validate := validator.New()
	validate.RegisterValidation("strong_password", utils.ValidPassword)

	app := service.NewAuthService(repo, redis, email, validate)
	authHandler := handlers.NewAuthServiceHandler(app)

	router := handlers.NewRouter(authHandler)

	port := config.ServerConfig.Port
	addr := config.ServerConfig.Host
	log.Printf("Starting server at port %s, address %s", port, addr)
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalln(err)
	}
}
