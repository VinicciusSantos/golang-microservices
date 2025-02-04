package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/VinicciusSantos/golang-microservices/authentication/data"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = 7070

var counts int64

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Println("Starting authentication service")

	conn := connectToDB()
	if conn == nil {
		log.Panic("Could not connect to Postgres")
	}

	app := Config{
		DB:     conn,
		Models: data.New(conn),
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", webPort),
		Handler: app.routes(),
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Panic("server failed to start:", err)
	}
}

func openDB(dsn string) (db *sql.DB, err error) {
	if db, err = sql.Open("pgx", dsn); err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return
}

func connectToDB() *sql.DB {
	const MAX_RETRIES = 10

	for {
		db, err := openDB(os.Getenv("DSN"))

		if err != nil {
			log.Println("Postgres not yet ready")
			counts++
		} else {
			log.Println("Connected to Postgres")
			return db
		}

		if counts > MAX_RETRIES {
			log.Println("Could not connect to Postgres")
			return nil
		}

		log.Println("backing off for two seconds")
		time.Sleep(2 * time.Second)
	}
}
