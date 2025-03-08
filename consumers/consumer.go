package consumers

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/11me/calef/config"
	"github.com/11me/calef/models"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
)

const binanceWsBaseUrl = "wss://stream.binance.com:9443/ws"

type BinanceConsumer struct {
	ctx context.Context

	log                      *slog.Logger
	conf                     *config.Config
	nc                       *nats.Conn
	tradesMu                 sync.Mutex
	ticksConnectionsBySymbol map[string]*websocket.Conn
	errCh                    chan error
}

func NewBinanceConsumer(ctx context.Context, nc *nats.Conn) *BinanceConsumer {
	return &BinanceConsumer{
		ctx:                      ctx,
		nc:                       nc,
		ticksConnectionsBySymbol: make(map[string]*websocket.Conn),
		log:                      slog.With("service", "BinanceConsumer"),
	}
}

func (c *BinanceConsumer) SubscribeTicks(symbols ...string) {
	for _, symbol := range symbols {
		c.ticksConnectionsBySymbol[symbol] = nil
	}
}

func (c *BinanceConsumer) Start() error {
	c.log.Info("starting binance consumer")
	c.connect()

	// TODO: wait all goroutines to finish.
	select {
	case <-c.ctx.Done():
		return nil
	case err := <-c.errCh:
		return err
	}
}

func (c *BinanceConsumer) connect() {
	for symbol := range c.ticksConnectionsBySymbol {
		go c.connectTrades(symbol)
	}
}

func (c *BinanceConsumer) connectTrades(symbol string) {
	c.tradesMu.Lock()
	defer c.tradesMu.Unlock()

	c.log.Debug(fmt.Sprintf("subscribing binance ticks for %q", symbol))

	url := fmt.Sprintf("%s/%s@trade", binanceWsBaseUrl, strings.ToLower(symbol))
	attempt := 0

	// Connect to stream.

	for {
		// TODO: listen for signal to quit
		conn, _, err := websocket.DefaultDialer.DialContext(c.ctx, url, nil)
		if err == nil {
			c.ticksConnectionsBySymbol[symbol] = conn
			attempt = 0

			break
		}

		attempt++

		c.log.Error("error dialing symbol", "symbol", symbol, "err", err, "attempt", attempt)

		if attempt > 3 {
			c.errCh <- fmt.Errorf("error dialing %q, all attempts failed: %w", symbol, err)

			return
		}

		time.Sleep(time.Duration(attempt) * time.Second)
	}

	// Start read loop

	go c.readTicks(symbol)
}

func (c *BinanceConsumer) readTicks(symbol string) {
	var conn *websocket.Conn

	c.tradesMu.Lock()
	conn = c.ticksConnectionsBySymbol[symbol]
	c.tradesMu.Unlock()

	// binance.ticks.btcusdt
	subj := models.BinanceTicksSubject + "." + symbol

	for {
		// TODO: listen for signal to quit

		_, msg, err := conn.ReadMessage()
		if err != nil {
			c.log.Error(fmt.Sprintf("error reading message for symbol %q: %v", symbol, err))

			conn.Close()

			go c.connectTrades(symbol)

			return
		}

		err = c.nc.Publish(subj, msg)
		if err != nil {
			c.log.Error("failed to publish message to subject", "subject", subj, "err", err)
		}
	}
}
