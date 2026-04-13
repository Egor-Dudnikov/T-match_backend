package main

import (
	"T-match_backend/configs"
	"T-match_backend/internal/cache"
	"T-match_backend/internal/handlers"
	"T-match_backend/internal/repository"
	"T-match_backend/internal/service"
	"T-match_backend/internal/utils"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

func main() {
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatalln("not JWT_SECRET in env")
	}
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	db, err := repository.PingDatabase(config.DbConfig)
	if err != nil {
		log.Fatalln(err)
	}

	dbr, err := cache.PingRedis(config.RedisConfig)
	if err != nil {
		log.Fatalln(err)
	}

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

	srv := &http.Server{
		Addr:         addr + port,
		Handler:      router,
		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}
	log.Printf("Starting server at port %s, address %s", port, addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Stop server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		log.Fatalln(err)
	}

	if err := db.Close(); err != nil {
		log.Println("DB close error:", err)
	}

	if err := dbr.Close(); err != nil {
		log.Println("Redis close error:", err)
	}

	log.Println("Server exited")
}
