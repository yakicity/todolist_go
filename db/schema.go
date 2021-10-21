package db

import (
	"time"
)

// Task corresponds to a row in `tasks` table
type Task struct {
	Id        uint64    `db:"id"`
	Title     string    `db:"title"`
	CreatedAt time.Time `db:"created_at"`
	IsDone    bool      `db:"is_done"`
}
