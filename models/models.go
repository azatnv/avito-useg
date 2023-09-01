package m

import (
	"time"
)

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
