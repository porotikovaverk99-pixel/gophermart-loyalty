-- migrations/000002_create_orders_table.up.sql
-- Создание таблицы orders
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    number VARCHAR(50) UNIQUE NOT NULL,
    status VARCHAR(15) NOT NULL DEFAULT 'NEW',
    accrual DECIMAL(10,2) DEFAULT 0,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_checked_at TIMESTAMP WITH TIME ZONE,
    next_check_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    retry_count INTEGER DEFAULT 0,
    backoff_factor INTEGER DEFAULT 1,
    
    CONSTRAINT valid_status CHECK (status IN ('NEW', 'PROCESSING', 'PROCESSED', 'INVALID')),
    CONSTRAINT positive_accrual CHECK (accrual >= 0)
);

-- Индексы
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_uploaded_at ON orders(uploaded_at DESC);
CREATE INDEX idx_orders_next_check ON orders(next_check_at) WHERE status IN ('NEW', 'PROCESSING');