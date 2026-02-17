-- migrations/000003_create_balance_transactions.up.sql
-- Создание таблицы balance_transactions
CREATE TABLE balance_transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(10) NOT NULL,
    order_number VARCHAR(50) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_type CHECK (type IN ('ACCRUAL', 'WITHDRAWAL')),
    CONSTRAINT positive_amount CHECK (amount > 0),
    CONSTRAINT unique_withdrawal_order UNIQUE (order_number, type) -- один номер на тип операции
);

-- Индексы
CREATE INDEX idx_transactions_user_id ON balance_transactions(user_id);
CREATE INDEX idx_transactions_processed_at ON balance_transactions(processed_at DESC);
CREATE INDEX idx_transactions_order_number ON balance_transactions(order_number);
CREATE INDEX idx_transactions_type ON balance_transactions(type);