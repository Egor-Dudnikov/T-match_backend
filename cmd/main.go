// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package main

import (
	"T-match_backend/configs"
	"T-match_backend/internal/cache"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/dadata"
	"T-match_backend/internal/handlers"
	"T-match_backend/internal/repository"
	"T-match_backend/internal/s3"
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
	_ "github.com/lib/pq"
)

func main() {
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatalln("not JWT_SECRET in env")
	}
	config := configs.LoadConfig()

	db, err := repository.PingDatabase(config.DbConfig)
	if err != nil {
		log.Fatalln(err)
	}

	dbr, err := cache.PingRedis(config.RedisConfig)
	if err != nil {
		log.Fatalln(err)
	}
	s3Client, err := s3.LoadS3(config.S3Config)
	if err != nil {
		log.Fatalln(err)
	}

	repo := repository.NewRepository(db)
	redis := cache.NewRedis(dbr)
	email := service.NewEmailClient(config.EmailConfig)
	s3Storage, err := s3.NewS3(s3Client, config.S3Config)
	if err != nil {
		log.Fatalln(err)
	}

	validate := validator.New()

	err = validate.RegisterValidation("strong_password", utils.ValidPassword)
	if err != nil {
		log.Fatalf("failed to register validation: %v", err)
	}
	err = validate.RegisterValidation("valid_role", utils.ValidRole)
	if err != nil {
		log.Fatalf("failed to register validation: %v", err)
	}

	dadataClient := dadata.NewClient()

	app := service.Newservice(repo, redis, email, validate, s3Storage, dadataClient)
	authHandler := handlers.NewServiceHandler(app, &config.CORSConfig)

	router := handlers.NewRouter(authHandler)

	port := config.ServerConfig.Port
	addr := config.ServerConfig.Host

	srv := &http.Server{
		Addr:         addr + port,
		Handler:      router,
		ReadTimeout:  constants.ServerReadTimeout,
		WriteTimeout: constants.ServerWriteTimeout,
		IdleTimeout:  constants.ServerIdleTimeout,
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

/*

Сурок (гофер) в стиле ASCII-арт - позаимствовано с
https://gist.github.com/belbomemo/b5e7dad10fa567a5fe8a

          ,_---~~~~~----._
   _,,_,*^____      _____``*g*\"*,
  / __/ /'     ^.  /      \ ^@q   f
 [  @f | @))    |  | @))   l  0 _/
  \`/   \~____ / __ \_____/    \
   |           _l__l_           I
   }          [______]           I
   ]            | | |            |
   ]             ~ ~             |
   |                            |
    |                           |

*/
