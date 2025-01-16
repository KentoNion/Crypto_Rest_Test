package server

import (
	"fmt"
	"github.com/shopspring/decimal"
)

type coinPriceTimeRequest struct {
	Coin      string `json:"coin"`
	Timestamp string `json:"timestamp"`
}

func (r *coinPriceTimeRequest) Validate() error {
	if r.Coin == "" {
		return fmt.Errorf("coin field cannot be empty")
	}
	if r.Timestamp == "" {
		return fmt.Errorf("timestamp field cannot be empty")
	}
	return nil
}

type coinPriceTimeResponse struct {
	Price     decimal.Decimal `json:"coin"`
	Timestamp string          `json:"timestamp"`
}
