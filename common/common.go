package common

import (
	"fmt"
	"strings"

	"github.com/11me/calef/models"
)

func BinanceTicksSubj(symbol string) string {
	symbol = strings.ToLower(strings.TrimSpace(symbol))
	return fmt.Sprintf("binancef.ticks.%s", symbol)
}

func BinanceBarsSubj(symbol string, tf models.Timeframe) string {
	symbol = strings.ToLower(strings.TrimSpace(symbol))
	return fmt.Sprintf("binancef.bars.%s.%s", tf.String(), symbol)
}
