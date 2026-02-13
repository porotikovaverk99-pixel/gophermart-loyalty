-- migrations/000004_alter_orders_accrual_default.down.sql
-- Откат: возвращаем значение по умолчанию 0
ALTER TABLE orders ALTER COLUMN accrual SET DEFAULT 0;

-- Возвращаем NULL обратно в 0 для новых заказов
UPDATE orders SET accrual = 0 WHERE accrual IS NULL AND status = 'NEW';