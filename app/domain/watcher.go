package domain

import (
	"context"
	"cryptoRestTest/internal/config"
	"github.com/shopspring/decimal"
	"log/slog"
	"strconv"
	"time"
)

type Watcher struct {
	store    CoinsStore
	log      *slog.Logger
	cfg      *config.Config
	provider Provider
	ctx      context.Context
}

func NewWatcher(ctx context.Context, store CoinsStore, log *slog.Logger, provider Provider, cfg *config.Config) *Watcher {
	return &Watcher{
		store:    store,
		log:      log,
		cfg:      cfg,
		provider: provider,
		ctx:      ctx,
	}
}

type CoinsStore interface {
	AddObserveredCoins(ctx context.Context, coins map[string]string) error
	GetObserveredCoinsList(ctx context.Context) (map[string]string, error)
	AddCoinsPrices(ctx context.Context, coins []Coin) error
	GetPrice(ctx context.Context, coin string, timestamp time.Time) (decimal.Decimal, time.Time, error)
	DeleteObserveredCoins(ctx context.Context, coins []string) error
}

type Provider interface {
	CoinsPrice(coins map[string]string) ([]Coin, error)
	VerifyCoins(coins []string) map[string]string
}

func (w Watcher) AddObserveredCoins(coins []string) error {
	const op = "domain.Watcher.AddObserveredCoins"

	verifiedCoins := w.provider.VerifyCoins(coins)

	if len(verifiedCoins) == 0 {
		w.log.Warn(op, "no coins to add to the store", ErrNoVerifiedCoins)
		return ErrNoVerifiedCoins
	}

	err := w.store.AddObserveredCoins(w.ctx, verifiedCoins)
	if err != nil {
		w.log.Error(op, "failed to add observered coins to store", err)
		return err
	}

	w.log.Debug(op, "added observered coins to store", verifiedCoins)
	return nil
}

func (w Watcher) GetObserveredCoinsList() ([]string, error) {
	const op = "domain.Watcher.GetObserveredCoinsList"
	w.log.Debug(op, "started GetObserveredCoinsList")

	coinsMap, err := w.store.GetObserveredCoinsList(w.ctx)
	if err != nil {
		w.log.Error(op, "failed to get observered coins list", err)
		return nil, err
	}
	coins := extractKeys(coinsMap)
	w.log.Debug(op, "got observered coins list: ", coins)
	return coins, nil
}

func (w Watcher) DeleteObserveredCoins(coins []string) error { //в этой функции я не преобразую []string в []Coin, тк не хочу получить лишний цикл
	const op = "domain.Watcher.DeleteObserveredCoins"
	w.log.Debug(op, "started DeleteObserveredCoins", coins)

	err := w.store.DeleteObserveredCoins(w.ctx, coins)
	if err != nil {
		w.log.Error(op, "failed to delete observered coins from store", err)
		return err
	}

	w.log.Debug(op, "successfully deleted observered coins")
	return nil
}

func (w Watcher) GetTimePrice(coin string, time time.Time) (decimal.Decimal, string, error) {
	const op = "domain.Watcher.GetLastPrice"

	w.log.Debug(op, "trying to get price for coin: ", coin, "time: ", time)
	price, time, err := w.store.GetPrice(w.ctx, string(coin), time)
	if err != nil {
		w.log.Error(op, "failed to get price for coin: ", coin, "time: ", time)
		return decimal.Zero, "", err
	}
	timestamp := strconv.FormatInt(time.Unix(), 10) //Перевод времени в изначальный формат который передавался в запросе
	w.log.Debug(op, "got price for coin: ", coin, "time: ", timestamp)
	return price, timestamp, nil
}

// функция которая будет пробегать по монетам записанных в список наблюдения (бд) и записывать их цену+время
func (w Watcher) ScanPrices() error {
	const op = "domain.Watcher.ScanPrices"

	coinsMap, err := w.store.GetObserveredCoinsList(w.ctx)
	if err != nil {
		w.log.Error(op, "failed to get observered coins list", err)
	}
	if coinsMap == nil || len(coinsMap) == 0 {
		w.log.Info(op, "no observered coins to scan", "0 coins in storage")
		return nil
	}
	w.log.Info(op, "starting ScanPrices for coins: ", coinsMap)
	coins, err := w.provider.CoinsPrice(coinsMap)
	if err != nil {
		w.log.Error(op, "failed to get coins prices", err)
		return err
	}
	err = w.store.AddCoinsPrices(w.ctx, coins)
	if err != nil {
		w.log.Error(op, "failed to add coins prices", err)
		return err
	}
	w.log.Info(op, "successfully added coins prices: ", extractKeys(coinsMap))
	return nil
}
