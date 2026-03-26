-- Payment Service schema
CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL UNIQUE,
    amount NUMERIC(10,2) NOT NULL CHECK (amount > 0),
    method TEXT NOT NULL DEFAULT 'cash',
    status TEXT NOT NULL,
    reference TEXT NOT NULL,
    paid_at TIMESTAMP NULL
);
