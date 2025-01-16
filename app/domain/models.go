package domain

import "errors"

type Coin string

var ErrNoVerifiedCoins = errors.New("no coins passed verification")
