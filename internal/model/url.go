package model

import "time"

type URL struct {
	ID          int64     `db:"id"`
	ShortCode   string    `db:"short_code"`
	OriginalURL string    `db:"original_url"`
	Clicks      int64     `db:"clicks"`
	CreatedAt   time.Time `db:"created_at"`
}
