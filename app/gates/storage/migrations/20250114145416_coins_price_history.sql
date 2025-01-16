-- +goose Up
ALTER DATABASE observered_coins SET timezone TO 'UTC';
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS observered_coins(
    coin VARCHAR(255) NOT NULL,
    time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    price NUMERIC NOT NULL,
    PRIMARY KEY (coin, time)
);
CREATE INDEX idx_observed_coins_time ON observered_coins(time);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS price_history;
-- +goose StatementEnd
