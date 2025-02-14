CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    login VARCHAR UNIQUE NOT NULL,
    password VARCHAR NOT NULL,
    balance INT NOT NULL CHECK (balance >= 0),
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    from_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    to_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    amount INT NOT NULL CHECK (amount > 0),
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS merch (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR UNIQUE NOT NULL,
    price INT NOT NULL CHECK (price > 0),
    is_selling BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE purchases (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    merch_id UUID REFERENCES merch(id) NOT NULL,
    price INT NOT NULL CHECK (price > 0),
    created_at TIMESTAMP DEFAULT now()
);

INSERT INTO merch (name, price) VALUES
    ('t-shirt', 80),
    ('cup', 20),
    ('book', 50),
    ('pen', 10),
    ('powerbank', 200),
    ('hoody', 300),
    ('umbrella', 200),
    ('socks', 10),
    ('wallet', 50),
    ('pink-hoody', 500);

CREATE INDEX idx_transactions_from_user ON transactions(from_user_id);
CREATE INDEX idx_transactions_to_user ON transactions(to_user_id);
CREATE INDEX idx_purchases_user ON purchases(user_id);
CREATE INDEX idx_merch_name ON merch(name);