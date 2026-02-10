package model

import "time"

type RequestAuth struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type User struct {
	ID           int64
	Login        string
	PasswordHash string
	CreatedAt    time.Time
}
