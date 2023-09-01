package h

import (
	"database/sql"
	"log"
	"net/http"
	"time"
)

var DB *sql.DB

var Timeout = time.Minute
var Infinity = time.Date(9999, 0, 0, 0, 0, 0, 0, time.UTC)

func UsersHandler(w http.ResponseWriter, r *http.Request) (err error) {
	switch r.Method {
	case http.MethodPost:
		err = createUserHandler(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	return
}

func SegmentsHandler(w http.ResponseWriter, r *http.Request) (err error) {
	switch r.Method {
	case http.MethodPost:
		err = createSegmentHandler(w, r)
	case http.MethodDelete:
		err = deleteSegmentHandler(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	return
}

func UserSegmentsHandler(w http.ResponseWriter, r *http.Request) (err error) {
	log.Println(r.URL.Path)
	switch r.Method {
	case http.MethodGet:
		err = readUserSegmentsHandler(w, r)
	case http.MethodPost:
		err = createUserSegmentsHandler(w, r)
	case http.MethodDelete:
		err = deleteUserSegmentsHandler(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	return
}
