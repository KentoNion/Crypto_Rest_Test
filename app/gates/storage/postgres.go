package storage

import (
	"context"
	"cryptoRestTest/domain"
	sq "github.com/Masterminds/squirrel"
	"github.com/bool64/sqluct"
	"github.com/jmoiron/sqlx"
	"log/slog"
)

func NewDB(db *sqlx.DB, log *slog.Logger) *Store {
	return &Store{
		db:  db,
		sm:  sqluct.Mapper{Dialect: sqluct.DialectPostgres},
		sq:  sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		log: log,
	}
}

func (s *Store) AddObserveredCoin(ctx context.Context, coin domain.Coin) error {
	const op = "gates.storage.AddObserveredCoin"
	s.log.Info(op, "trying to add coin : ", coin)

	query := s.sq.Insert("observered_coins").
		Columns("coin_name").
		Values(coin).
		Suffix("ON CONFLICT DO NOTHING")
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
	if rows, _ := rows.RowsAffected(); rows == 0 {
		s.log.Error(op, "no rows affected for coin: ", coin)
		return ErrNoRowsAffected
	}
	s.log.Info(op, "added coin: ", coin)
	return nil
}

func (s *Store) DeleteObserveredCoin(ctx context.Context, coin domain.Coin) error {
	const op = "gates.storage.DeleteObserveredCoin"
	s.log.Info(op, "trying to delete coin : ", coin)

	query := s.sq.Delete("observered_coins").
		Where(sq.Eq{"coin_name": coin})
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
	if rows, _ := rows.RowsAffected(); rows == 0 {
		s.log.Error(op, "no rows affected for coin: ", coin)
		return ErrNoRowsAffected
	}
	s.log.Info(op, "deleted coin: ", coin)
	return nil
}

func (s *Store) GetObserveredCoinsList(ctx context.Context) ([]domain.Coin, error) {
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
	var coins []domain.Coin
	err = s.db.SelectContext(ctx, &coins, qry, args...)
	if err != nil {
		s.log.Error(op, "failed to execute query", err)
		return nil, err
	}
	s.log.Info(op, "successfully retrieved observered coins list")
	return coins, nil
}
