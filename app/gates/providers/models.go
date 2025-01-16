package coingecko

import (
	"fmt"
)

var ErrEmptyPriceCurrency = fmt.Errorf("No price found for this currency")
var ErrCoinDontExist = fmt.Errorf("Could not find this coin")
