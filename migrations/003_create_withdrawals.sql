-- +goose Up
CREATE TABLE IF NOT EXISTS withdrawals (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id),
    order_number VARCHAR(255) NOT NULL UNIQUE,
    sum          NUMERIC(12,2) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS withdrawals_user_id_idx ON withdrawals(user_id);

-- +goose Down
DROP TABLE IF EXISTS withdrawals;
