package main

import (
	"log"
	"net/http"
	"time"
)

var timeout = time.Minute

func wrapError(h func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte(err.Error())); err != nil {
				return
			}
			log.Println(err)
		}
	})
}

func segmentHandler(w http.ResponseWriter, r *http.Request) (err error) {
	switch r.Method {
	case http.MethodPost:
		err = execSqlOnSegments(w, r, createSegmentSQL)
	case http.MethodDelete:
		err = execSqlOnSegments(w, r, deleteSegmentSQL)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	return
}

func userHandler(w http.ResponseWriter, r *http.Request) (err error) {
	switch r.Method {
	case http.MethodGet:
		err = getUserSegments(w, r)
	case http.MethodPost:
		err = addUserSegments(w, r)
	case http.MethodDelete:
		err = deleteUserSegments(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	return
}
