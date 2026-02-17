// Package model содержит структуры данных, используемые во всем приложении.
package model

import "time"

// BalanceTransaction представляет операцию начисления или списания баллов.
type BalanceTransaction struct {
	ID          int64     `db:"id" json:"-"`                      // внутренний идентификатор
	UserID      int64     `db:"user_id" json:"-"`                 // идентификатор пользователя
	Type        string    `db:"type" json:"type"`                 // ACCRUAL или WITHDRAWAL
	OrderNumber string    `db:"order_number" json:"order"`        // номер заказа
	Amount      float64   `db:"amount" json:"amount"`             // сумма
	ProcessedAt time.Time `db:"processed_at" json:"processed_at"` // время операции
}

// WithdrawalResponse — модель списания в системе лояльности..
type Withdrawal struct {
	Order       string    `json:"order"`        // номер заказа
	Amount      float64   `json:"sum"`          // сумма списания
	ProcessedAt time.Time `json:"processed_at"` // время операции
}

// BalanceResponse — ответ с текущим балансом и суммой списаний.
type BalanceResponse struct {
	Current   float64 `json:"current"`   // текущий баланс пользователя
	Withdrawn float64 `json:"withdrawn"` // сумма списанных баллов
}

// WithdrawRequest — запрос на списание баллов.
type WithdrawRequest struct {
	Order string  `json:"order"` // номер заказа для оплаты
	Sum   float64 `json:"sum"`   // сумма списания
}
