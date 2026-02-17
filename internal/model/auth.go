// Package model содержит структуры данных, используемые во всем приложении.
package model

import "time"

// RequestAuth — тело запроса для регистрации и аутентификации.
type RequestAuth struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// User — модель пользователя в системе.
type User struct {
	ID           int64     `db:"id"`            // идентификатор пользователя
	Login        string    `db:"login"`         // уникальный логин
	PasswordHash string    `db:"password_hash"` // bcrypt-хэш пароля
	CreatedAt    time.Time `db:"created_at"`    // дата регистрации
}
