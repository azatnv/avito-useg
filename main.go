package main

import (
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

	db, err = sql.Open("pgx", fmt.Sprintf("postgres://%s:%s@localhost:5432/%s", u, p, dbName))
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(db)

	mux := http.NewServeMux()

	mux.Handle("/users", wrapError(usersHandler))
	mux.Handle("/segments", wrapError(segmentsHandler))
	mux.Handle("/users/segments", wrapError(userSegmentsHandler))

	err = http.ListenAndServe(":80", http.TimeoutHandler(mux, timeout, "time limit exceeded"))
	if err != nil {
		log.Fatal("server error")
	}
}
