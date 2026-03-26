-- Order Service schema
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    menu_id INT NOT NULL,
    qty INT NOT NULL CHECK (qty > 0),
    status TEXT NOT NULL
);
