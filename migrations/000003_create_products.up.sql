CREATE TABLE products (
    id          BIGSERIAL PRIMARY KEY,
    owner_id    BIGINT        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name        VARCHAR(200)  NOT NULL,
    description VARCHAR(2000) NOT NULL DEFAULT '',
    price       BIGINT        NOT NULL CHECK (price >= 0),
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_products_owner_id ON products (owner_id);
