-- migrations/000004_alter_orders_accrual_default.up.sql
-- Изменяем значение по умолчанию для поля accrual с 0 на NULL
ALTER TABLE orders ALTER COLUMN accrual SET DEFAULT NULL;

-- Обновляем существующие записи: для новых заказов accrual должен быть NULL
UPDATE orders SET accrual = NULL WHERE accrual = 0 AND status = 'NEW';