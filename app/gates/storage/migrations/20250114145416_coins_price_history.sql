-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS observered_coins(
    coin VARCHAR(255) NOT NULL,
    time TIMESTAMP NOT NULL DEFAULT NOW(),
    price NUMERIC NOT NULL,
    CONSTRAINT fk_coin FOREIGN KEY (coin) REFERENCES observered_coins (coin) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS price_history;
-- +goose StatementEnd
