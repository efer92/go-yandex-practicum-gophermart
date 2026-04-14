-- +goose Up
CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE IF NOT EXISTS orders (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    number      VARCHAR(255) NOT NULL UNIQUE,
    status      order_status NOT NULL DEFAULT 'NEW',
    accrual     NUMERIC(12,2),
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS orders_user_id_idx ON orders(user_id);
CREATE INDEX IF NOT EXISTS orders_pending_idx ON orders(status)
    WHERE status NOT IN ('INVALID', 'PROCESSED');

-- +goose Down
DROP TABLE IF EXISTS orders;
DROP TYPE IF EXISTS order_status;
