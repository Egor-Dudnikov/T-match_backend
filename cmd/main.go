// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package main

import (
	"T-match_backend/configs"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/handlers"
	"T-match_backend/internal/service"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatalln("not JWT_SECRET in env")
	}

	config := configs.LoadConfig()
	app, err := service.RegService(config)
	if err != nil {
		log.Fatalln(err)
	}

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

	if err := app.CloseDB(); err != nil {
		log.Println("DB close error:", err)
	}

	if err := app.CloseRedis(); err != nil {
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
