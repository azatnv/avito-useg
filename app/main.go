package main

import (
	"avito-useg/handlers"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func main() {
	err := godotenv.Load(".env.db")
	if err != nil {
		log.Fatal("error loading .env file")
	}
	u := os.Getenv("POSTGRES_USER")
	p := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	h.DB, err = sql.Open("pgx", fmt.Sprintf("postgres://%s:%s@db:5432/%s", u, p, dbName))
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(h.DB)

	mux := http.NewServeMux()

	mux.Handle("/users", h.WrapError(h.UsersHandler))
	mux.Handle("/segments", h.WrapError(h.SegmentsHandler))
	mux.Handle("/users/segments", h.WrapError(h.UserSegmentsHandler))

	err = http.ListenAndServe(":80", http.TimeoutHandler(mux, h.Timeout, "time limit exceeded"))
	if err != nil {
		log.Fatal("server error")
	}
}
