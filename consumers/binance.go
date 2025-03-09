package consumers

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"github.com/11me/calef/config"
	"github.com/11me/calef/models"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/valyala/fastjson"
)

const (
	binanceWsBaseUrl = "wss://stream.binance.com:9443/ws"
)

type BinanceConsumer struct {
	ctx     context.Context
	conf    *config.Config
	log     *slog.Logger
	nc      *nats.Conn
	conn    *websocket.Conn
	symbols []string
	errCh   chan error
}

func NewBinanceConsumer(ctx context.Context, nc *nats.Conn) *BinanceConsumer {
	return &BinanceConsumer{
		ctx:   ctx,
		nc:    nc,
		log:   slog.With("service", "BinanceConsumer"),
		errCh: make(chan error),
	}

}

func (c *BinanceConsumer) Start() error {
	c.log.Info("starting binance consumer")

	conn, err := c.connect()
	if err != nil {
		return err
	}

	c.conn = conn

	c.log.Info("connected to binance")

	go c.reconnect()
	go c.readTicks()

	// TODO: wait all goroutines to finish.
	select {
	case <-c.ctx.Done():
		return nil
	}
}

func (c *BinanceConsumer) connect() (*websocket.Conn, error) {
	c.log.Info(fmt.Sprintf("connecting to binance %q", binanceWsBaseUrl))

	conn, _, err := websocket.DefaultDialer.DialContext(c.ctx, binanceWsBaseUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to dial binance: %w", err)
	}

	params := make([]string, 0, len(c.symbols))
	for _, symbol := range c.symbols {
		params = append(params, symbol+"@aggTrade")
	}

	if len(params) > 0 {
		err := conn.WriteJSON(map[string]any{
			"method": "SUBSCRIBE",
			"params": params,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to subscribe to tickers on binance: %w", err)
		}
	}

	return conn, nil
}

func (c *BinanceConsumer) reconnect() {
	for {
		err := <-c.errCh

		if err == nil {
			// App shutdown, errCh was closed.
			select {
			case <-c.ctx.Done():
				return
			default: // Proceed reconnection loop.
			}
		}

		c.log.Error(fmt.Sprintf("binance connection was closed with error: %v, reconnection", err))

		var conn *websocket.Conn

		for {
			var connErr error

			conn, connErr = c.connect()
			if connErr != nil {
				timeout := time.Second * time.Duration(1+rand.Intn(3))
				c.log.Error(fmt.Sprintf("error reconnection to binance: %v, next retry in %v", connErr, timeout))

				select {
				case <-c.ctx.Done():
					return
				case <-time.After(timeout):
				}

				continue
			}

			// Success.
			break
		}

		c.conn = conn

		c.log.Info("Succesfully reconnected to binance")

		go c.readTicks()
	}
}

func (c *BinanceConsumer) readTicks() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			c.log.Error(fmt.Sprintf("error reading message for symbol: %v", err))
			c.errCh <- err

			return
		}

		parser := fastjson.Parser{}

		val, err := parser.ParseBytes(msg)
		if err != nil {
			c.log.Error("failed to parse json", "value", string(msg), "err", err)

			continue
		}

		res := val.Get("result")
		if res != nil {
			// Ignore response for subscription.
			continue
		}

		symbol := val.Get("s")
		subj := models.BinanceTicksSubject + "." + strings.ToLower(string(symbol.GetStringBytes()))

		err = c.nc.Publish(subj, msg)
		if err != nil {
			c.log.Error("failed to publish message to subject", "subject", subj, "err", err)
		}
	}
}

func (c *BinanceConsumer) SubscribeTicks(symbols ...string) {
	c.symbols = append(c.symbols, symbols...)
}
