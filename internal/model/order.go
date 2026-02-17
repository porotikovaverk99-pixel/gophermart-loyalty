// Package model содержит структуры данных, используемые во всем приложении.
package model

import "time"

// Order — модель заказа в системе лояльности.
type Order struct {
	ID         int64     `db:"id" json:"-"`                      // внутренний идентификатор
	UserID     int64     `db:"user_id" json:"-"`                 // владелец заказа
	Number     string    `db:"number" json:"number"`             // номер заказа
	Status     string    `db:"status" json:"status"`             // NEW, PROCESSING, PROCESSED, INVALID
	Accrual    *float64  `db:"accrual" json:"accrual,omitempty"` // начисленные баллы
	UploadedAt time.Time `db:"uploaded_at" json:"uploaded_at"`   // время загрузки

	// Поля для внутреннего использования (не возвращаются в API)
	LastCheckedAt *time.Time `db:"last_checked_at" json:"-"` // последняя проверка статуса
	NextCheckAt   *time.Time `db:"next_check_at" json:"-"`   // планируемая следующая проверка
	RetryCount    int        `db:"retry_count" json:"-"`     // счетчик повторов при ошибках
}

// AccrualTask — задача для асинхронного начисления баллов.
type AccrualTask struct {
	UserID   int64   // получатель баллов
	OrderNum string  // номер заказа
	Amount   float64 // сумма начисления
}
