package model

import "time"

type Order struct {
	ID         int64     `db:"id" json:"-"`
	UserID     int64     `db:"user_id" json:"-"`
	Number     string    `db:"number" json:"number"`
	Status     string    `db:"status" json:"status"`
	Accrual    *float64  `db:"accrual" json:"accrual,omitempty"`
	UploadedAt time.Time `db:"uploaded_at" json:"uploaded_at"`

	LastCheckedAt *time.Time `db:"last_checked_at" json:"-"`
	NextCheckAt   *time.Time `db:"next_check_at" json:"-"`
	RetryCount    int        `db:"retry_count" json:"-"`
}
