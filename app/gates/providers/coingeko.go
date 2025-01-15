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
}

func NewClient(cfg *config.Config, log *slog.Logger) *Client {
	return &Client{
		cfg: cfg,
		log: log,
		cg:  api.NewDefaultClient(),
	}
}

func (c Client) OneCoinPrice(ctx context.Context, coin string) (decimal.Decimal, error) {
	const op = "gates.providers.coingecko.OneCoinPrice"

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

func (c Client) CoinsPrice(ctx context.Context, coins []string) (map[string]decimal.Decimal, error) {
	const op = "gates.providers.coingecko.CoinsPrice"

	c.log.Info(op, "trying to get prices for coins: ", coins)
	currency := c.cfg.CoinsWatcher.Currency

	coinsQuery := strings.Join(coins, ",") //API требует передачи списка монет как string с запятыми
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
