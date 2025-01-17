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

func (s *Store) AddObserveredCoins(ctx context.Context, coins map[string]string) error {
	const op = "gates.storage.AddObserveredCoins"
	s.log.Debug(op, "trying to add coins:", coins)

	query := s.sq.Insert("observered_coins").
		Columns("coin", "id").
		Suffix("ON CONFLICT DO NOTHING")

	for coin, id := range coins {
		query = query.Values(coin, id)
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

	s.log.Debug(op, "successfully added coins:", coins)
	return nil
}

func (s *Store) DeleteObserveredCoins(ctx context.Context, coins []string) error {
	const op = "gates.storage.DeleteObserveredCoin"
	s.log.Debug(op, "trying to delete coins:", coins)

	query := s.sq.Delete("observered_coins").
		Where(sq.Eq{"coin": coins})
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

	s.log.Debug(op, "successfully deleted coins:", coins)
	return nil
}

func (s *Store) GetObserveredCoinsList(ctx context.Context) (map[string]string, error) {
	const op = "gates.storage.GetObserveredCoinsList"
	s.log.Debug(op, "trying to get observered coins list")

	query := s.sq.Select("coin", "id").
		From("observered_coins")
	qry, args, err := query.ToSql()
	s.log.Debug(op, "query: ", qry, "args: ", args)
	if err != nil {
		s.log.Error(op, "failed to build query", err)
		return nil, err
	}

	type coinRow struct {
		Coin string `db:"coin"`
		ID   string `db:"id"`
	}

	var rows []coinRow
	err = s.db.SelectContext(ctx, &rows, qry, args...)
	if err != nil {
		s.log.Error(op, "failed to execute query", err)
		return nil, err
	}

	// Перегоняем в мапу
	coins := make(map[string]string, len(rows))
	for _, row := range rows {
		coins[row.Coin] = row.ID
	}

	s.log.Debug(op, "successfully retrieved observered coins list")
	return coins, nil
}

func (s *Store) AddCoinsPrices(ctx context.Context, coins []domain.Coin) error {
	const op = "gates.storage.AddCoinsPrices"
	s.log.Debug(op, "trying to add coin prices")

	// Начинаем построение запроса
	query := s.sq.Insert("price_history").
		Columns("coin", "price", "time").
		Suffix("ON CONFLICT DO NOTHING")

	for _, coin := range coins {
		query = query.Values(coin.Name, coin.Price, time.Now().UTC())
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

	s.log.Debug(op, "successfully added coin prices")
	return nil
}

func (s *Store) GetPrice(ctx context.Context, coin string, timestamp time.Time) (decimal.Decimal, time.Time, error) {
	const op = "gates.storage.GetPrice"
	s.log.Debug(op, "trying to get price for coin", "coin", coin, "time", timestamp)

	query := s.sq.Select("price, time").
		From("price_history").
		Where(sq.Eq{"coin": coin}).
		OrderBy("ABS(EXTRACT(EPOCH FROM (time - ?)))"). // Используем значение времени для сортировки
		Limit(1)

	qry, args, err := query.ToSql()
	s.log.Debug(op, "query: ", qry, "args: ", args)
	if err != nil {
		s.log.Error(op, "failed to build query", err)
		return decimal.Zero, time.Time{}, err
	}

	// Порядок аргументов: сначала coin, затем timestamp
	args = append(args, timestamp)

	var r struct {
		Price decimal.Decimal `db:"price"`
		Time  time.Time       `db:"time"`
	}
	err = s.db.GetContext(ctx, &r, qry, args...)
	if err != nil {
		s.log.Error(op, "failed to execute query", err)
		return decimal.Zero, time.Time{}, err
	}

	s.log.Debug(op, "successfully retrieved price",
		"coin", coin,
		"request_timestamp", timestamp,
		"found_timestamp", r.Time,
		"price", r.Price)
	return r.Price, r.Time, nil
}
