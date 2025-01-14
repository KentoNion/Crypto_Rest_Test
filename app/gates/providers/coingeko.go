package coingecko

import (
	"context"
	"cryptoRestTest/internal/config"
	"fmt"
	"github.com/JulianToledano/goingecko/v3/api" //ООоочень простой в использовании package специально под coingecko
	"github.com/shopspring/decimal"
	"log/slog"
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
	const op = "gates.providers.coingecko.Price"

	c.log.Info(op, "trying to get price of coin: ", coin)
	currency := c.cfg.CoinsWatcher.Currency
	price, err := c.cg.SimplePrice(ctx, coin, currency, false)
	if err != nil {
		c.log.Error(op, "Error getting price from coingecko", err)
		return decimal.Zero, err
	}
	if coinPrice, ok := price[coin][currency]; ok {
		c.log.Info(op, "got price for coin: ", coin)
		return decimal.NewFromFloat(coinPrice), ErrEmptyPriceCurrency
	}

	// Если цена не найдена
	err = fmt.Errorf("price for coin %s not found", coin)
	c.log.Error(op, "no price for currency : ", currency)
	c.log.Warn(op, "perhaps this currency does not exist?", currency)
	return decimal.Zero, ErrEmptyPriceCurrency
}
