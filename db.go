package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
const selectAllUsersSQL = "SELECT id FROM users"

const addUserSegmentsSQL = "INSERT INTO user2seg(u_id, s_id, date_add, date_end) VALUES ($1, $2, $3, $4)"
const deleteUserSegmentsSQL = "DELETE FROM user2seg WHERE u_id=($1) and s_id=($2)"
const selectUserSegmentsSQL = `
	SELECT name FROM user2seg as u2s
	JOIN segments as s ON u2s.s_id = s.id 
	WHERE u_id=($1)`

func createSegment(ctx context.Context, conn *sql.Conn, name string) (err error) {
	_, err = conn.ExecContext(ctx, createSegmentSQL, strings.ToUpper(name))
	return
}

func checkSegment(ctx context.Context, conn *sql.Conn, name string) (int, bool, error) {
	segmentID, err := selectSegmentID(ctx, conn, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return segmentID, false, nil
		}
		return segmentID, false, err
	}
	return segmentID, true, nil
}

func selectSegmentID(ctx context.Context, conn *sql.Conn, name string) (int, error) {
	var segmentID int
	err := conn.QueryRowContext(ctx, selectSegmentSQL, strings.ToUpper(name)).Scan(&segmentID)
	if err != nil {
		return -1, err
	}
	return segmentID, nil
}

func assignSegmentForUsers(ctx context.Context, conn *sql.Conn, name string, uids []int) (err error) {
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

	segmentID, err := selectSegmentID(ctx, conn, name)
	if err != nil {
		return err
	}

	for _, uid := range uids {
		_, err = tx.ExecContext(ctx, addUserSegmentsSQL, uid, segmentID, time.Now(), infinity)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return
}

func assignSegmentsForUser(ctx context.Context, conn *sql.Conn, us UserSegments) (err error) {
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
	return
}

func deleteUserSegments(ctx context.Context, conn *sql.Conn, us UserSegments) (err error) {
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
	return
}

func selectUserSegments(ctx context.Context, conn *sql.Conn, uid int) (res []Segment, err error) {
	rows, err := conn.QueryContext(ctx, selectUserSegmentsSQL, uid)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var s Segment
		if err := rows.Scan(&s.Name); err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	return
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

func selectAllUsers(ctx context.Context, conn *sql.Conn) (ids []int, err error) {
	rows, err := conn.QueryContext(ctx, selectAllUsersSQL)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var i int
		if err = rows.Scan(&i); err != nil {
			return nil, err
		}
		ids = append(ids, i)
	}
	return ids, err
}
