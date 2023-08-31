package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var timeout = time.Minute
var infinity = time.Date(9999, 0, 0, 0, 0, 0, 0, time.UTC)

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

func usersHandler(w http.ResponseWriter, r *http.Request) (err error) {
	switch r.Method {
	case http.MethodPost:
		err = createUserHandler(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	return
}

func segmentsHandler(w http.ResponseWriter, r *http.Request) (err error) {
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

func userSegmentsHandler(w http.ResponseWriter, r *http.Request) (err error) {
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

func createUserHandler(w http.ResponseWriter, r *http.Request) error {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(r.Body)
	var data bytes.Buffer
	if _, err := io.Copy(&data, r.Body); err != nil {
		return errors.New(fmt.Sprintf("JSON reading failed: %s", err))
	}

	var users []User
	if err := json.Unmarshal(data.Bytes(), &users); err != nil {
		return errors.New(fmt.Sprintf("JSON unmarshaling failed: %s", err))
	}

	ctx := r.Context()
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func(conn *sql.Conn) {
		err := conn.Close()
		if err != nil {
			return
		}
	}(conn)

	for _, u := range users {
		err = createUser(ctx, conn, u.UserID)
	}
	w.WriteHeader(http.StatusOK)
	return nil
}

func createSegmentHandler(w http.ResponseWriter, r *http.Request) error {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(r.Body)
	var data bytes.Buffer
	if _, err := io.Copy(&data, r.Body); err != nil {
		return errors.New(fmt.Sprintf("JSON reading failed: %s", err))
	}

	var s Segment
	if err := json.Unmarshal(data.Bytes(), &s); err != nil {
		return errors.New(fmt.Sprintf("JSON unmarshaling failed: %s", err))
	}
	if s.Name == "" {
		return errors.New("there is no segment")
	}

	ctx := r.Context()
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func(conn *sql.Conn) {
		err := conn.Close()
		if err != nil {
			return
		}
	}(conn)

	// create segment
	err = createSegment(ctx, conn, s.Name)
	if err != nil {
		return err
	}

	p, err := strconv.Atoi(r.URL.Query().Get("percent"))
	if err == nil {
		// assign segment to a percentage of users
		ids, err := selectAllUsers(ctx, conn)
		if err != nil {
			return err
		}
		rand.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })
		n := p * len(ids) / 100
		if n == 0 {
			return errors.New("not enough users for that percent")
		}
		err = assignSegmentForUsers(ctx, conn, s.Name, ids[:n])
		if err != nil {
			return err
		}
	}
	w.WriteHeader(http.StatusOK)
	return nil
}

func deleteSegmentHandler(w http.ResponseWriter, r *http.Request) error {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(r.Body)
	var data bytes.Buffer
	if _, err := io.Copy(&data, r.Body); err != nil {
		return errors.New(fmt.Sprintf("JSON reading failed: %s", err))
	}

	var s Segment
	if err := json.Unmarshal(data.Bytes(), &s); err != nil {
		return errors.New(fmt.Sprintf("JSON unmarshaling failed: %s", err))
	}
	if s.Name == "" {
		return errors.New("there is no segment")
	}

	ctx := r.Context()
	_, err := db.ExecContext(ctx, deleteSegmentSQL, strings.ToUpper(s.Name))
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	return nil
}

func createUserSegmentsHandler(w http.ResponseWriter, r *http.Request) error {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(r.Body)
	var data bytes.Buffer
	if _, err := io.Copy(&data, r.Body); err != nil {
		return errors.New(fmt.Sprintf("JSON reading failed: %s", err))
	}

	var us UserSegments
	if err := json.Unmarshal(data.Bytes(), &us); err != nil {
		return errors.New(fmt.Sprintf("JSON unmarshaling failed: %s", err))
	}
	if us.DateEnd.IsZero() {
		us.DateEnd = infinity
	}
	if us.UserID == 0 || us.Segments == nil {
		return errors.New("bad request")
	}

	// common connection
	ctx := r.Context()
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func(conn *sql.Conn) {
		err := conn.Close()
		if err != nil {
			return
		}
	}(conn)

	// create new user if it doesn't exist
	err = checkOrCreateUser(ctx, conn, us.UserID)
	if err != nil {
		return err
	}

	// add all segments for that user
	// if at least one segment does not exist or
	// some unique error occurs (in user2seg table) then the request is rejected
	err = assignSegmentsForUser(ctx, conn, us)
	w.WriteHeader(http.StatusOK)
	return nil
}

func readUserSegmentsHandler(w http.ResponseWriter, r *http.Request) error {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(r.Body)
	var data bytes.Buffer
	if _, err := io.Copy(&data, r.Body); err != nil {
		return errors.New(fmt.Sprintf("JSON reading failed: %s", err))
	}

	var u User
	if err := json.Unmarshal(data.Bytes(), &u); err != nil {
		return errors.New(fmt.Sprintf("JSON unmarshaling failed: %s", err))
	}
	if u.UserID == 0 {
		return errors.New("bad request")
	}

	ctx := r.Context()
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func(conn *sql.Conn) {
		err := conn.Close()
		if err != nil {
			return
		}
	}(conn)

	res, err := selectUserSegments(ctx, conn, u.UserID)
	if err != nil {
		return err
	}

	resBytes, err := json.Marshal(res)
	if err != nil {
		return err
	}
	_, err = w.Write(resBytes)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	return nil
}

func deleteUserSegmentsHandler(w http.ResponseWriter, r *http.Request) error {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(r.Body)
	var data bytes.Buffer
	if _, err := io.Copy(&data, r.Body); err != nil {
		return errors.New(fmt.Sprintf("JSON reading failed: %s", err))
	}

	var us UserSegments
	if err := json.Unmarshal(data.Bytes(), &us); err != nil {
		return errors.New(fmt.Sprintf("JSON unmarshaling failed: %s", err))
	}
	if us.UserID == 0 || us.Segments == nil {
		return errors.New("bad request")
	}

	// common connection
	ctx := r.Context()
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func(conn *sql.Conn) {
		err := conn.Close()
		if err != nil {
			return
		}
	}(conn)

	// check user
	b, err := checkUser(ctx, conn, us.UserID)
	if err != nil {
		return err
	}
	if !b {
		return errors.New(fmt.Sprintf("there is no such user %d", us.UserID))
	}

	// remove existing segments for that user
	// unknown segments are skipped
	err = deleteUserSegments(ctx, conn, us)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
