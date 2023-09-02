package db

import (
	"avito-useg/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var infinity = time.Date(9999, 0, 0, 0, 0, 0, 0, time.UTC)

const selectUserSQL = "SELECT id FROM users WHERE id=($1)"
const createUserSQL = "INSERT INTO users(id) VALUES ($1)"
const selectAllUsersSQL = "SELECT id FROM users"

const addUserSegmentsSQL = "INSERT INTO user2seg(u_id, s_id, date_add, date_end) VALUES ($1, $2, $3, $4)"
const deleteUserSegmentsSQL = "DELETE FROM user2seg WHERE u_id=($1) and s_id=($2)"
const selectUserSegmentsSQL = `
	SELECT name FROM user2seg as u2s
	JOIN segments as s ON u2s.s_id = s.id 
	WHERE u_id=($1)`

func AssignSegmentForUsers(ctx context.Context, conn *sql.Conn, name string, ids []int) (err error) {
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

	for _, uid := range ids {
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

func AssignSegmentsForUser(ctx context.Context, conn *sql.Conn, us m.UserSegments) (err error) {
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

func DeleteUserSegments(ctx context.Context, conn *sql.Conn, us m.UserSegments) (err error) {
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

func SelectUserSegments(ctx context.Context, conn *sql.Conn, uid int) (res []m.Segment, err error) {
	rows, err := conn.QueryContext(ctx, selectUserSegmentsSQL, uid)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var s m.Segment
		if err := rows.Scan(&s.Name); err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	return
}

func CheckOrCreateUser(ctx context.Context, conn *sql.Conn, userID int) (err error) {
	b, err := CheckUser(ctx, conn, userID)
	if err != nil {
		return err
	}
	if !b {
		if err = CreateUser(ctx, conn, userID); err != nil {
			return err
		}
	}
	return
}

func CheckUser(ctx context.Context, conn *sql.Conn, userID int) (bool, error) {
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

func CreateUser(ctx context.Context, conn *sql.Conn, userID int) (err error) {
	_, err = conn.ExecContext(ctx, createUserSQL, userID)
	return
}

func SelectAllUsers(ctx context.Context, conn *sql.Conn) (ids []int, err error) {
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
