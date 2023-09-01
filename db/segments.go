package db

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

const selectSegmentSQL = "SELECT id FROM segments WHERE name=($1)"
const createSegmentSQL = "INSERT INTO segments(name) VALUES ($1)"
const deleteSegmentSQL = "DELETE FROM segments WHERE name=($1)"

func CreateSegment(ctx context.Context, conn *sql.Conn, name string) (err error) {
	_, err = conn.ExecContext(ctx, createSegmentSQL, strings.ToUpper(name))
	return
}

func DeleteSegment(ctx context.Context, conn *sql.Conn, name string) (err error) {
	_, err = conn.ExecContext(ctx, deleteSegmentSQL, strings.ToUpper(name))
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
