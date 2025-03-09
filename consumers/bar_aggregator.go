package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/11me/calef/models"
	"github.com/nats-io/nats.go"
	"github.com/valyala/fastjson"
)

type BarAggregator struct {
	ctx        context.Context
	nc         *nats.Conn
	log        *slog.Logger
	symbols    []string
	timeframes []time.Duration
}

func NewBarAggregator(ctx context.Context, nc *nats.Conn) *BarAggregator {
	return &BarAggregator{
		ctx: ctx,
		nc:  nc,
		log: slog.With("service", "BarAggregator"),
	}
}

func (ba *BarAggregator) AddTimeframes(tf ...time.Duration) {
	ba.timeframes = append(ba.timeframes, tf...)
}

func (ba *BarAggregator) AddSymbols(symbols ...string) {
	ba.symbols = append(ba.symbols, symbols...)
}

func (ba *BarAggregator) Start() error {
	for _, tf := range ba.timeframes {
		for _, symbol := range ba.symbols {
			go ba.startAggregate(tf, symbol)
		}
	}

	<-ba.ctx.Done()

	return nil
}

func (ba *BarAggregator) startAggregate(tf time.Duration, symbol string) {
	ba.log.Info(fmt.Sprintf("start aggregator for %s %s", symbol, tf))

	subjTicks := models.BinanceTicksSubject + "." + symbol

	sub, err := ba.nc.SubscribeSync(subjTicks)
	if err != nil {
		ba.log.Error(fmt.Sprintf("failed to subscribe to %s for symbol %q", subjTicks, symbol))
		return
	}
	defer sub.Drain()

	// Publish bars on subject "binancef.bars.<tf>.<symbol>" -> "binancef.bars.1m.btcusdt"
	subjBars := fmt.Sprintf("binancef.bars.%s.%s", intervalToStr(tf), symbol)

	var currentBar *models.Bar
	parser := fastjson.Parser{}

	for {
		msg, err := sub.NextMsg(5 * time.Second)
		if err != nil {
			ba.log.Error("error reading next message", "err", err)

			continue
		}

		val, err := parser.ParseBytes(msg.Data)
		if err != nil {
			ba.log.Error("failed to parse tick", "err", err, "data", string(msg.Data))

			continue
		}

		symbolTick := string(val.GetStringBytes("s"))
		// If the tick doesn't match our symbol, skip.
		if strings.ToLower(symbolTick) != symbol {

			continue
		}

		priceStr := string(val.GetStringBytes("p"))
		quantityStr := string(val.GetStringBytes("q"))
		eventTime := val.GetInt64("E")

		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			ba.log.Error("failed to parse price", "err", err)

			continue
		}

		quantity, err := strconv.ParseFloat(quantityStr, 64)
		if err != nil {
			ba.log.Error("failed to parse quantity", "err", err)

			continue
		}

		tickTime := time.Unix(0, eventTime*int64(time.Millisecond))
		tickBucket := tickTime.Truncate(tf)

		// If there's no current candle or the tick is in a new bucket, finalize and start a new candle.
		if currentBar == nil || tickBucket.After(currentBar.StartTime) {
			// Finalize the existing candle, if any.
			if currentBar != nil {
				currentBar.IsClosed = true
				if err := ba.publishBar(subjBars, currentBar); err != nil {
					ba.log.Error("failed to publish finalized candle", "subject", subjBars, "err", err)
				}
			}

			// Start a new candle.
			currentBar = &models.Bar{
				Open:      price,
				High:      price,
				Low:       price,
				Close:     price,
				Volume:    quantity,
				StartTime: tickBucket,
				IsClosed:  false,
			}
		} else {
			// Update the current candle.
			if price > currentBar.High {
				currentBar.High = price
			}

			if price < currentBar.Low {
				currentBar.Low = price
			}

			currentBar.Close = price
			currentBar.Volume += quantity
		}

		if err := ba.publishBar(subjBars, currentBar); err != nil {
			ba.log.Error("failed to publish candle", "subject", subjBars, "err", err)
		}
	}
}

func (ba *BarAggregator) publishBar(subj string, candle *models.Bar) error {
	data, err := json.Marshal(candle)
	if err != nil {
		return err
	}
	return ba.nc.Publish(subj, data)
}

func intervalToStr(d time.Duration) string {
	// Weeks
	if d >= 7*24*time.Hour {
		weeks := d / (7 * 24 * time.Hour)
		return fmt.Sprintf("%dw", weeks)
	}

	// Days
	if d >= 24*time.Hour {
		days := d / (24 * time.Hour)
		return fmt.Sprintf("%dd", days)
	}

	// Hours
	if d >= time.Hour {
		hours := d / time.Hour
		return fmt.Sprintf("%dh", hours)
	}

	// Minutes
	if d >= time.Minute {
		minutes := d / time.Minute
		return fmt.Sprintf("%dm", minutes)
	}

	// Seconds
	if d >= time.Second {
		seconds := d / time.Second
		return fmt.Sprintf("%ds", seconds)
	}

	return d.String()
}
