package coingecko

import (
	"context"
	"cryptoRestTest/domain"
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
func (c Client) VerifyCoins(coins []string) map[string]string {
	const op = "gates.providers.coingecko.VerifyCoins"

	ctx, cancel := context.WithTimeout(c.ctx, c.cfg.CoinsWatcher.Timeout)
	defer cancel()

	c.log.Info(op, "Fetching coin list to verify:", coins)
	list, err := c.cg.CoinsList(ctx)
	if err != nil {
		c.log.Error(op, "Error fetching coin list from Coingecko", err)
		return nil
	}

	// Преобразуем список монет от CoinGecko в map для ускорения поиска
	coinMap := make(map[string]string, len(list))
	for _, coin := range list {
		coinMap[coin.Symbol] = coin.ID // Используем символы (symbol) для сопоставления
	}

	verifiedCoins := make(map[string]string)
	for _, coin := range coins {
		// Преобразуем входной символ в ID
		if id, exists := coinMap[coin]; exists {
			verifiedCoins[coin] = id
			c.log.Debug(op, "Coin verified:", coin)
		} else {
			c.log.Warn(op, "Coin not found in CoinGecko list:", coin)
		}
	}

	c.log.Info(op, "Verified coins:", verifiedCoins)
	return verifiedCoins
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
func (c Client) CoinsPrice(coins map[string]string) ([]domain.Coin, error) {
	const op = "gates.providers.coingecko.CoinsPrice"

	ctx, cancel := context.WithTimeout(c.ctx, c.cfg.CoinsWatcher.Timeout)
	defer cancel()

	c.log.Info(op, "trying to get prices for coins:", coins)
	currency := c.cfg.CoinsWatcher.Currency

	// Собираем значения из мапы coins (id) в строку через запятую
	coinIDs := make([]string, 0, len(coins))
	for _, id := range coins {
		coinIDs = append(coinIDs, id)
	}
	coinsQuery := strings.Join(coinIDs, ",")

	// Получаем цены через API CoinGecko
	priceMap, err := c.cg.SimplePrice(ctx, coinsQuery, currency, false)
	if err != nil {
		c.log.Error(op, "Error getting prices from coingecko", err)
		return nil, err
	}

	// Формируем результат в формате []domain.Coin
	var result []domain.Coin
	for name, id := range coins {
		if cgPrices, exists := priceMap[id]; exists {
			if price, ok := cgPrices[currency]; ok {
				coin := domain.Coin{
					Name:  name,
					Id:    id,
					Price: decimal.NewFromFloat(price),
				}
				result = append(result, coin)
			} else {
				c.log.Warn(op, "Price not found for id in the specified currency:", id, currency)
			}
		} else {
			c.log.Warn(op, "ID not found in CoinGecko price map:", id)
		}
	}

	if len(result) == 0 {
		err := fmt.Errorf("no prices found for the provided coins")
		c.log.Error(op, "no prices found", err)
		return nil, err
	}

	c.log.Debug(op, "retrieved coin prices:", result)
	c.log.Info(op, "successfully retrieved prices for coins")
	return result, nil
}
