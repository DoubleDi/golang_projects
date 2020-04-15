DO
$$
BEGIN
    CREATE TYPE transaction_state AS ENUM ('win', 'lost');
    CREATE TABLE transactions (
        id VARCHAR NOT NULL PRIMARY KEY,
        state transaction_state NOT NULL,
        amount NUMERIC (10, 2) NOT NULL CHECK (amount >= 0),
        created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
        reverted BOOLEAN NOT NULL DEFAULT FALSE
    );
    CREATE INDEX created_at_idx ON transactions USING btree (created_at DESC);
    CREATE TABLE balance (
        amount NUMERIC (10, 2) NOT NULL CHECK (amount >= 0),
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );
    INSERT INTO balance VALUES (0, NOW());
END;
$$;
