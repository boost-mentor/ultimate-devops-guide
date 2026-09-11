CREATE TABLE IF NOT EXISTS orders (
    id          BIGSERIAL PRIMARY KEY,
    target      TEXT        NOT NULL,
    price_usd   INTEGER     NOT NULL CHECK (price_usd >= 0),
    status      TEXT        NOT NULL,
    handled_by  TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS orders_created_at_idx ON orders (created_at DESC);
