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
	"math/rand"
	"net/http"
	"strconv"
)

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

	var s m.Segment
	if err := json.Unmarshal(data.Bytes(), &s); err != nil {
		return errors.New(fmt.Sprintf("JSON unmarshaling failed: %s", err))
	}
	if s.Name == "" {
		return errors.New("there is no segment")
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

	// create segment
	err = db.CreateSegment(ctx, conn, s.Name)
	if err != nil {
		return err
	}

	p, err := strconv.Atoi(r.URL.Query().Get("percent"))
	if err == nil {
		// assign segment to a percentage of users
		ids, err := db.SelectAllUsers(ctx, conn)
		if err != nil {
			return err
		}
		rand.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })
		n := p * len(ids) / 100
		if n == 0 {
			return errors.New("not enough users for that percent")
		}
		err = db.AssignSegmentForUsers(ctx, conn, s.Name, ids[:n])
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

	var s m.Segment
	if err := json.Unmarshal(data.Bytes(), &s); err != nil {
		return errors.New(fmt.Sprintf("JSON unmarshaling failed: %s", err))
	}
	if s.Name == "" {
		return errors.New("there is no segment")
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

	err = db.DeleteSegment(ctx, conn, s.Name)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	return nil
}
