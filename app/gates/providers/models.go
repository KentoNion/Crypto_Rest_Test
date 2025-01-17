package coingecko

import (
	"fmt"
)

var ErrEmptyPriceCurrency = fmt.Errorf("No price found for this currency")
var ErrCoinDontExist = fmt.Errorf("Could not find this coin")

func getMapValues(m map[string]string) []string {
	values := make([]string, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}
