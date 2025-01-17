package domain

import (
	"errors"
	"github.com/shopspring/decimal"
)

var ErrNoVerifiedCoins = errors.New("no coins passed verification")

type Coin struct {
	Name  string
	Id    string
	Price decimal.Decimal
}

func extractKeys(input map[string]string) []string {
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	return keys
}
