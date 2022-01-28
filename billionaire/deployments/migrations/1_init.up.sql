CREATE TABLE wallet_hour_history (
    datetime TIMESTAMPTZ NOT NULL,
    amount NUMERIC NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX ON wallet_hour_history (datetime);