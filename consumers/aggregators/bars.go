package aggregators

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"
	"time"

	c "github.com/11me/calef/common"
	"github.com/11me/calef/consumers"
	"github.com/11me/calef/models"
	"github.com/nats-io/nats.go"
	"github.com/valyala/fastjson"
)

type BarAggregator struct {
	ctx        context.Context
	nc         *nats.Conn
	log        *slog.Logger
	symbol     string
	tf         models.Timeframe
	consumer   *consumers.Consumer
	currentBar *models.Bar
}

func NewBarAggregator(ctx context.Context, nc *nats.Conn, symbol string, tf models.Timeframe) *BarAggregator {
	agg := &BarAggregator{
		ctx:      ctx,
		nc:       nc,
		symbol:   symbol,
		tf:       tf,
		log:      slog.With("service", "BarAggregator", "symbol", symbol, "timeframe", tf.String()),
		consumer: consumers.NewConsumer(ctx, nc),
	}

	agg.consumer.
		SetConcurrency(1).
		SetLogger(agg.log).
		Subscribe(c.BinanceTicksSubj(agg.symbol), agg)

	return agg
}

func (ba *BarAggregator) Spawn() error {
	return ba.consumer.Start()
}

func (ba *BarAggregator) Stop() error {
	return ba.consumer.Stop()
}

func (ba *BarAggregator) Handle(msg *nats.Msg) error {
	parser := fastjson.Parser{}
	val, err := parser.ParseBytes(msg.Data)
	if err != nil {
		ba.log.Error("failed to parse tick", "err", err, "data", string(msg.Data))
		return err
	}

	symbolTick := string(val.GetStringBytes("s"))
	if strings.ToLower(symbolTick) != ba.symbol {
		return nil
	}

	priceStr := string(val.GetStringBytes("p"))
	quantityStr := string(val.GetStringBytes("q"))
	eventTime := val.GetInt64("E")

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		ba.log.Error("failed to parse price", "err", err)
		return err
	}

	quantity, err := strconv.ParseFloat(quantityStr, 64)
	if err != nil {
		ba.log.Error("failed to parse quantity", "err", err)
		return err
	}

	// Determine the bucket for the current tick based on the timeframe.
	tickTime := time.Unix(0, eventTime*int64(time.Millisecond))
	tickBucket := tickTime.Truncate(time.Duration(ba.tf))

	subjBars := c.BinanceBarsSubj(ba.symbol, ba.tf)

	if ba.currentBar == nil || tickBucket.After(ba.currentBar.StartTime) {
		if ba.currentBar != nil {
			ba.currentBar.IsClosed = true
			if err := ba.publishBar(subjBars, ba.currentBar); err != nil {
				ba.log.Error("failed to publish finalized candle", "subject", subjBars, "err", err)
			}
		}

		ba.currentBar = &models.Bar{
			Symbol:    ba.symbol,
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
		if price > ba.currentBar.High {
			ba.currentBar.High = price
		}

		if price < ba.currentBar.Low {
			ba.currentBar.Low = price
		}

		ba.currentBar.Close = price
		ba.currentBar.Volume += quantity
	}

	if err := ba.publishBar(subjBars, ba.currentBar); err != nil {
		ba.log.Error("failed to publish candle", "subject", subjBars, "err", err)
		return err
	}

	return nil
}

func (ba *BarAggregator) publishBar(subj string, candle *models.Bar) error {
	data, err := json.Marshal(candle)
	if err != nil {
		return err
	}

	return ba.nc.Publish(subj, data)
}
