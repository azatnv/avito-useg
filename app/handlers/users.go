package h

import (
	"avito-useg/db"
	"avito-useg/models"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

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

	var users []m.User
	if err := json.Unmarshal(data.Bytes(), &users); err != nil {
		return errors.New(fmt.Sprintf("JSON unmarshaling failed: %s", err))
	}

	ctx := r.Context()
	conn, err := DB.Conn(ctx)
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
		err = db.CreateUser(ctx, conn, u.UserID)
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

	var us m.UserSegments
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
	conn, err := DB.Conn(ctx)
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
	err = db.CheckOrCreateUser(ctx, conn, us.UserID)
	if err != nil {
		return err
	}

	// add all segments for that user
	// if at least one segment does not exist or
	// some unique error occurs (in user2seg table) then the request is rejected
	err = db.AssignSegmentsForUser(ctx, conn, us)
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

	var u m.User
	if err := json.Unmarshal(data.Bytes(), &u); err != nil {
		return errors.New(fmt.Sprintf("JSON unmarshaling failed: %s", err))
	}
	if u.UserID == 0 {
		return errors.New("bad request")
	}

	ctx := r.Context()
	conn, err := DB.Conn(ctx)
	if err != nil {
		return err
	}
	defer func(conn *sql.Conn) {
		err := conn.Close()
		if err != nil {
			return
		}
	}(conn)

	res, err := db.SelectUserSegments(ctx, conn, u.UserID)
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

	var us m.UserSegments
	if err := json.Unmarshal(data.Bytes(), &us); err != nil {
		return errors.New(fmt.Sprintf("JSON unmarshaling failed: %s", err))
	}
	if us.UserID == 0 || us.Segments == nil {
		return errors.New("bad request")
	}

	// common connection
	ctx := r.Context()
	conn, err := DB.Conn(ctx)
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
	b, err := db.CheckUser(ctx, conn, us.UserID)
	if err != nil {
		return err
	}
	if !b {
		return errors.New(fmt.Sprintf("there is no such user %d", us.UserID))
	}

	// remove existing segments for that user
	// unknown segments are skipped
	err = db.DeleteUserSegments(ctx, conn, us)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
