package coingecko

import (
	"context"
	"cryptoRestTest/internal/config"
	"fmt"
	"github.com/JulianToledano/goingecko/v3/api" //ООоочень простой в использовании package специально под coingecko
	"github.com/shopspring/decimal"
	"log/slog"
	"strings"
)

type Client struct {
	cg  *api.Client
	cfg *config.Config
	log *slog.Logger
	ctx context.Context
}

func NewClient(ctx context.Context, cfg *config.Config, log *slog.Logger) *Client {
	return &Client{
		cfg: cfg,
		log: log,
		cg:  api.NewDefaultClient(),
		ctx: ctx,
	}
}

// функция проверяет монету на наличие (существование) на coingecko
func (c Client) VerifyCoin(coin string) error {
	const op = "gates.providers.coingecko.VerifyCoin"

	ctx, cancel := context.WithTimeout(c.ctx, c.cfg.CoinsWatcher.Timeout)
	defer cancel()

	// Получаем список всех монет через API Coingecko
	c.log.Info(op, "Fetching coin list to verify:", coin)
	coins, err := c.cg.CoinsList(ctx)
	if err != nil {
		c.log.Error(op, "Error fetching coin list from Coingecko", err)
		return ErrCoinDontExist
	}

	// Проверяем, есть ли монета в списке
	for _, i := range coins {
		if i.ID == coin {
			c.log.Info(op, "Coin verified:", coin)
			return nil
		}
	}

	// Монета не найдена
	c.log.Warn(op, "Coin not found:", coin)
	return nil
}

/* в конечном итоге не пригодилось
// Получает цену для одной монеты
func (c Client) OneCoinPrice(coin string) (decimal.Decimal, error) {
	const op = "gates.providers.coingecko.OneCoinPrice"

	ctx, cancel := context.WithTimeout(c.ctx, c.cfg.CoinsWatcher.Timeout)
	defer cancel()

	c.log.Info(op, "trying to get price of coin: ", coin)
	currency := c.cfg.CoinsWatcher.Currency
	price, err := c.cg.SimplePrice(ctx, coin, currency, false)
	if err != nil {
		c.log.Error(op, "Error getting price from coingecko", err)
		return decimal.Zero, err
	}
	if coinPrice, ok := price[coin][currency]; ok {
		c.log.Info(op, "got price for coin: ", coin)
		return decimal.NewFromFloat(coinPrice), nil //-----------корректный выход из функции
	}

	// Если цена не найдена
	err = fmt.Errorf("price for coin %s not found", coin)
	c.log.Error(op, "no price for currency : ", currency)
	c.log.Warn(op, "perhaps this currency does not exist?", currency)
	return decimal.Zero, ErrEmptyPriceCurrency
}
*/

// получает слайс монет - отдаёт мапу монета-цена
func (c Client) CoinsPrice(coins []string) (map[string]decimal.Decimal, error) {
	const op = "gates.providers.coingecko.CoinsPrice"

	ctx, cancel := context.WithTimeout(c.ctx, c.cfg.CoinsWatcher.Timeout)
	defer cancel()

	c.log.Info(op, "trying to get prices for coins: ", coins)
	currency := c.cfg.CoinsWatcher.Currency

	coinsQuery := strings.Join(coins, ",")
	priceMap, err := c.cg.SimplePrice(ctx, coinsQuery, currency, false)
	if err != nil {
		c.log.Error(op, "Error getting prices from coingecko", err)
		return nil, err
	}

	// Формируем результат
	result := make(map[string]decimal.Decimal)
	for _, coin := range coins {
		if coinPrice, ok := priceMap[coin][currency]; ok {
			result[coin] = decimal.NewFromFloat(coinPrice)
		} else {
			c.log.Warn(op, "price for coin not found: ", coin)
		}
	}

	if len(result) == 0 {
		err := fmt.Errorf("no prices found for the provided coins")
		c.log.Error(op, "no prices found", err)
		return nil, err
	}

	c.log.Info(op, "successfully retrieved prices for coins")
	return result, nil
}
