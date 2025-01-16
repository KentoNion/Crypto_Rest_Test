package storage

import (
	"context"
	"cryptoRestTest/domain"
	sq "github.com/Masterminds/squirrel"
	"github.com/bool64/sqluct"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	"log/slog"
	"time"
)

func NewDB(db *sqlx.DB, log *slog.Logger) *Store {
	return &Store{
		db:  db,
		sm:  sqluct.Mapper{Dialect: sqluct.DialectPostgres},
		sq:  sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		log: log,
	}
}

func (s *Store) AddObserveredCoins(ctx context.Context, coins []domain.Coin) error {
	const op = "gates.storage.AddObserveredCoin"
	s.log.Info(op, "trying to add coins:", coins)

	// Проверяем, есть ли монеты для добавления
	if len(coins) == 0 {
		s.log.Warn(op, "no coins provided for insertion")
		return nil
	}

	// Формируем SQL-запрос
	query := s.sq.Insert("observered_coins").
		Columns("coin_name").
		Suffix("ON CONFLICT DO NOTHING")

	// Добавляем все монеты в запрос
	for _, coin := range coins {
		query = query.Values(coin)
	}
	qry, args, err := query.ToSql()
	s.log.Debug(op, "query: ", qry, "args: ", args)
	if err != nil {
		s.log.Error(op, "failed to build query", err)
		return err
	}

	rows, err := s.db.ExecContext(ctx, qry, args...)
	if err != nil {
		s.log.Error(op, "failed to execute query", err)
		return err
	}

	if rowsAffected, _ := rows.RowsAffected(); rowsAffected == 0 {
		s.log.Warn(op, "no rows were inserted for coins:", coins)
		return ErrNoRowsAffected
	}

	s.log.Info(op, "successfully added coins:", coins)
	return nil
}

func (s *Store) DeleteObserveredCoins(ctx context.Context, coins []string) error {
	const op = "gates.storage.DeleteObserveredCoin"
	s.log.Info(op, "trying to delete coins:", coins)

	query := s.sq.Delete("observered_coins").
		Where(sq.Eq{"coin_name": coins}) // sq.Eq поддерживает массив значений
	qry, args, err := query.ToSql()
	s.log.Debug(op, "query: ", qry, "args: ", args)
	if err != nil {
		s.log.Error(op, "failed to build query", err)
		return err
	}

	rows, err := s.db.ExecContext(ctx, qry, args...)
	if err != nil {
		s.log.Error(op, "failed to execute query", err)
		return err
	}

	if rowsAffected, _ := rows.RowsAffected(); rowsAffected == 0 {
		s.log.Warn(op, "no rows were deleted for coins:", coins)
		return ErrNoRowsAffected
	}

	s.log.Info(op, "successfully deleted coins:", coins)
	return nil
}

func (s *Store) GetObserveredCoinsList(ctx context.Context) ([]string, error) {
	const op = "gates.storage.GetObserveredCoinsList"
	s.log.Info(op, "trying to get observered coins list")

	query := s.sq.Select("coin_name").
		From("observered_coins")
	qry, args, err := query.ToSql()
	s.log.Debug(op, "query: ", qry, "args: ", args)
	if err != nil {
		s.log.Error(op, "failed to build query", err)
		return nil, err
	}
	var coins []string
	err = s.db.SelectContext(ctx, &coins, qry, args...)
	if err != nil {
		s.log.Error(op, "failed to execute query", err)
		return nil, err
	}
	s.log.Info(op, "successfully retrieved observered coins list")
	return coins, nil
}

func (s *Store) AddCoinsPrices(ctx context.Context, prices map[string]decimal.Decimal) error {
	const op = "gates.storage.AddCoinsPrices"
	s.log.Info(op, "trying to add coin prices")

	// Начинаем построение запроса
	query := s.sq.Insert("price_history").
		Columns("coin", "price", "time").
		Suffix("ON CONFLICT DO NOTHING")

	for coin, price := range prices {
		query = query.Values(coin, price, nil) // nil во время тк в таблице NOW()
	}

	// Генерируем SQL-запрос
	qry, args, err := query.ToSql()
	s.log.Debug(op, "query: ", qry, "args: ", args)
	if err != nil {
		s.log.Error(op, "failed to build query", err)
		return err
	}

	rows, err := s.db.ExecContext(ctx, qry, args...)
	if err != nil {
		s.log.Error(op, "failed to execute query", err)
		return err
	}

	// Проверяем, были ли затронуты строки
	if rowsAffected, _ := rows.RowsAffected(); rowsAffected == 0 {
		s.log.Error(op, "no rows affected")
		return ErrNoRowsAffected
	}

	s.log.Info(op, "successfully added coin prices")
	return nil
}

func (s *Store) GetPrice(ctx context.Context, coin string, timestamp time.Time) (decimal.Decimal, time.Time, error) {
	const op = "gates.storage.GetPrice"
	s.log.Info(op, "trying to get price for coin", "coin", coin, "timestamp", timestamp)

	query := s.sq.Select("price, time").
		From("observered_coins").
		Where(sq.Eq{"coin": coin}).
		OrderBy("ABS(EXTRACT(EPOCH FROM (time - ?)))").
		Limit(1) // Находим запись с минимальной разницей во времени

	qry, args, err := query.ToSql()
	s.log.Debug(op, "query: ", qry, "args: ", args)
	if err != nil {
		s.log.Error(op, "failed to build query", err)
		return decimal.Zero, time.Time{}, err
	}
	args = append([]interface{}{timestamp}, args...)

	var r priceTime
	err = s.db.GetContext(ctx, &r, qry, args...)
	if err != nil {
		s.log.Error(op, "failed to execute query", err)
		return decimal.Zero, time.Time{}, err
	}

	s.log.Info(op, "successfully retrieved price",
		"coin", coin,
		"request_timestamp", timestamp,
		"found_timestamp", r.Time,
		"price", r.Price)
	return r.Price, r.Time, nil
}
