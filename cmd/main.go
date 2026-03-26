package main

import (
	"T-match_backend/configs"
	"T-match_backend/internal/http"
	"T-match_backend/internal/rw"
	"log"
	stdhttp "net/http"

	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

func main() {
	config, err := configs.PingConfig()
	if err != nil {
		log.Fatalln(err)
	}

	db, err := rw.PingDatabase(config.DbConfig)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	dbr, err := rw.RedisPing(config.RedisConfig)
	if err != nil {
		log.Fatalln(err)
	}
	defer dbr.Close()

	app := &http.App{
		Db:  db,
		Dbr: dbr,
		Log: log.Default(),
		Cfg: config,
	}

	router := http.NewRouter(app)

	port := config.ServerConfig.Port
	addr := config.ServerConfig.Host
	log.Printf("Starting server at port %s, address %s", port, addr)
	if err := stdhttp.ListenAndServe(port, router); err != nil {
		log.Fatalln(err)
	}
}
