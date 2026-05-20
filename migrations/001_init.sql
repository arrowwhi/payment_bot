CREATE TABLE IF NOT EXISTS calendar_cache (
    year       INT         NOT NULL,
    month      INT         NOT NULL,
    days_data  TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (year, month)
);

CREATE TABLE IF NOT EXISTS users (
    telegram_id BIGINT      PRIMARY KEY,
    salary      NUMERIC(15,2) NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
