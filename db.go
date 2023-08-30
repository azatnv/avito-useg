package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var db *sql.DB

type User struct {
	UserID int `json:"id"`
}

type Segment struct {
	Name string `json:"name"`
}

type UserSegments struct {
	User
	Segments []Segment `json:"segments"`
	DateEnd  time.Time `json:"date_end"`
}

const selectSegmentSQL = "SELECT id FROM segments WHERE name=($1)"
const createSegmentSQL = "INSERT INTO segments(name) VALUES ($1)"
const deleteSegmentSQL = "DELETE FROM segments WHERE name=($1)"

const selectUserSQL = "SELECT id FROM users WHERE id=($1)"
const createUserSQL = "INSERT INTO users(id) VALUES ($1)"

const addUserSegmentsSQL = "INSERT INTO user2seg(u_id, s_id, date_add, date_end) VALUES ($1, $2, $3, $4)"
const deleteUserSegmentsSQL = "DELETE FROM user2seg WHERE u_id=($1) and s_id=($2)"
const selectUserSegmentsSQL = `
	SELECT name FROM user2seg as u2s
	JOIN segments as s ON u2s.s_id = s.id 
	WHERE u_id=($1)`

func execSqlOnSegments(w http.ResponseWriter, r *http.Request, stm string) error {
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
	if _, err := db.ExecContext(ctx, stm, strings.ToUpper(s.Name)); err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func addUserSegments(w http.ResponseWriter, r *http.Request) error {
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
		us.DateEnd = time.Date(9999, 0, 0, 0, 0, 0, 0, time.UTC)
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
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			return
		}
	}(tx)

	// if at least one segment does not exist or any unique error then the request is rejected
	for _, s := range us.Segments {
		segmentID, b, err := checkSegment(ctx, conn, s.Name)
		if err != nil {
			return err
		}
		if !b {
			return errors.New(fmt.Sprintf("there is no such segment %s", s))
		}
		_, err = tx.ExecContext(ctx, addUserSegmentsSQL, us.UserID, segmentID, time.Now(), us.DateEnd)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func checkSegment(ctx context.Context, conn *sql.Conn, name string) (int, bool, error) {
	var segmentID int
	err := conn.QueryRowContext(ctx, selectSegmentSQL, strings.ToUpper(name)).Scan(&segmentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return segmentID, false, nil
		}
		return segmentID, false, err
	}
	return segmentID, true, nil
}

func checkOrCreateUser(ctx context.Context, conn *sql.Conn, userID int) (err error) {
	b, err := checkUser(ctx, conn, userID)
	if err != nil {
		return err
	}
	if !b {
		if err = createUser(ctx, conn, userID); err != nil {
			return err
		}
	}
	return
}

func checkUser(ctx context.Context, conn *sql.Conn, userID int) (bool, error) {
	var selectedID int
	err := conn.QueryRowContext(ctx, selectUserSQL, userID).Scan(&selectedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func createUser(ctx context.Context, conn *sql.Conn, userID int) (err error) {
	_, err = conn.ExecContext(ctx, createUserSQL, userID)
	return
}

func getUserSegments(w http.ResponseWriter, r *http.Request) error {
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
	rows, err := db.QueryContext(ctx, selectUserSegmentsSQL, u.UserID)
	if err != nil {
		return err
	}

	var a []Segment
	for rows.Next() {
		var s Segment
		if err := rows.Scan(&s.Name); err != nil {
			return err
		}
		a = append(a, s)
	}

	aBytes, err := json.Marshal(a)
	if err != nil {
		return err
	}
	_, err = w.Write(aBytes)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	return nil
}

func deleteUserSegments(w http.ResponseWriter, r *http.Request) error {
	// if user exists then delete his segments
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
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			return
		}
	}(tx)

	// unknown segments are skipped
	for _, s := range us.Segments {
		segmentID, b, err := checkSegment(ctx, conn, s.Name)
		if err != nil {
			return err
		}
		if !b {
			continue
		}
		_, err = tx.ExecContext(ctx, deleteUserSegmentsSQL, us.UserID, segmentID)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
