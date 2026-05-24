CREATE TABLE IF NOT EXISTS transactions (
    id           BIGSERIAL PRIMARY KEY,
    order_id     BIGINT       NULL REFERENCES orders(id),
    order_number VARCHAR(255) NULL,
    user_id      BIGINT       NOT NULL REFERENCES users(id),
    sum          REAL PRECISION NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
