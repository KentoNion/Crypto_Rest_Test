package domain

import (
	"context"
	coingecko "cryptoRestTest/gates/providers"
	"cryptoRestTest/internal/config"
	"github.com/shopspring/decimal"
	"log/slog"
	"strconv"
	"sync"
	"time"
)

type Watcher struct {
	store    CoinsStore
	log      *slog.Logger
	cfg      *config.Config
	provider *coingecko.Client
	ctx      context.Context
}

func NewWatcher(ctx context.Context, store CoinsStore, log *slog.Logger, cfg *config.Config) *Watcher {
	return &Watcher{
		store:    store,
		log:      log,
		cfg:      cfg,
		provider: coingecko.NewClient(ctx, cfg, log),
		ctx:      ctx,
	}
}

type CoinsStore interface {
	AddObserveredCoins(ctx context.Context, coins []Coin) error
	GetObserveredCoinsList(ctx context.Context) ([]string, error)
	AddCoinsPrices(ctx context.Context, prices map[string]decimal.Decimal) error
	GetPrice(ctx context.Context, coin string, timestamp time.Time) (decimal.Decimal, time.Time, error)
	DeleteObserveredCoins(ctx context.Context, coins []string) error
}

func (w Watcher) AddObserveredCoins(coins []string) error {
	const op = "domain.Watcher.AddObserveredCoins"

	var mu sync.Mutex
	var verifiedCoins []Coin
	w.log.Debug(op, "started AddObserveredCoins", coins)

	wg := sync.WaitGroup{}
	for _, coin := range coins {
		wg.Add(1)
		go func(coin Coin) {
			defer wg.Done() // Уменьшаем счетчик WaitGroup по завершению горутины

			err := w.provider.VerifyCoin(string(coin))
			if err != nil {
				w.log.Debug(op, "failed to verify coin", err)
				return
			}

			mu.Lock()
			verifiedCoins = append(verifiedCoins, coin)
			mu.Unlock()
		}(Coin(coin)) // Передаем `coin` в замыкание, чтобы избежать проблем с захватом переменной
	}
	wg.Wait()

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

	coins, err := w.store.GetObserveredCoinsList(w.ctx)
	if err != nil {
		w.log.Error(op, "failed to get observered coins list", err)
		return nil, err
	}
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

func (w Watcher) GetTimePrice(coin Coin, time time.Time) (decimal.Decimal, string, error) {
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

	coins, err := w.store.GetObserveredCoinsList(w.ctx)
	if err != nil {
		w.log.Error(op, "failed to get observered coins list", err)
	}
	w.log.Info(op, "starting ScanPrices for coins: ", coins)
	coinPrices, err := w.provider.CoinsPrice(coins)
	if err != nil {
		w.log.Error(op, "failed to get coins prices", err)
		return err
	}
	err = w.store.AddCoinsPrices(w.ctx, coinPrices)
	if err != nil {
		w.log.Error(op, "failed to add coins prices", err)
		return err
	}
	w.log.Info(op, "successfully added coins prices: ", coinPrices)
	return nil
}
