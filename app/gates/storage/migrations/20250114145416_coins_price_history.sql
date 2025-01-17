-- +goose Up
ALTER DATABASE coins SET timezone TO 'UTC';
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS price_history(
    coin VARCHAR(255) NOT NULL,
    time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    price NUMERIC NOT NULL,
    PRIMARY KEY (coin, time)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS price_history;
-- +goose StatementEnd
