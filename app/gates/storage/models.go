package storage

import (
	"errors"
	sq "github.com/Masterminds/squirrel"
	"github.com/bool64/sqluct"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	"log/slog"
	"time"
)

type Store struct {
	db  *sqlx.DB
	sq  sq.StatementBuilderType
	sm  sqluct.Mapper
	log *slog.Logger
}

var ErrNoRowsAffected = errors.New("no rows affected")

// Функция для вычисления абсолютной разницы во времени
func absDuration(t1, t2 time.Time) time.Duration {
	if t1.Before(t2) {
		return t2.Sub(t1)
	}
	return t1.Sub(t2)
}

// структура для получения цены\времени
type priceTime struct {
	Price decimal.Decimal `db:"price"`
	Time  time.Time       `db:"time"`
}
