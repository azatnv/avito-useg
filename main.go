package main

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"net/http"
)

func main() {
	var err error
	db, err = sql.Open("pgx",
		"postgres://azatnv:joipjdfJSFDJidj@localhost:5432/segments")
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

	mux.Handle("/segments", wrapError(segmentHandler))
	mux.Handle("/users/segments", wrapError(userHandler))

	err = http.ListenAndServe(":8080", http.TimeoutHandler(mux, timeout, "time limit exceeded"))
	if err != nil {
		panic(err)
	}
}
