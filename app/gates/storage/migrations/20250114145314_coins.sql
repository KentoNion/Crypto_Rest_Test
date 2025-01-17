-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS observered_coins(
    coin VARCHAR(255) PRIMARY KEY,
    id VARCHAR(255) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP Table observered_coins;
-- +goose StatementEnd
