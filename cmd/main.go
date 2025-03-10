package main

import (
	"context"
	"log"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/11me/calef/config"
	"github.com/11me/calef/consumers"
	"github.com/nats-io/nats.go"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	slog.SetLogLoggerLevel(slog.LevelDebug)

	conf, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	nc, err := nats.Connect(conf.Nats.URL)
	if err != nil {
		log.Fatal(err)
	}

	consumer := consumers.NewBinanceConsumer(ctx, nc)
	consumer.SubscribeTicks("btcusdt", "ethusdt")

	barAggr := consumers.NewBarAggregator(ctx, nc)
	barAggr.AddSymbols("btcusdt", "ethusdt")
	barAggr.AddTimeframes(time.Second, time.Minute, 5*time.Minute, 15*time.Minute)

	go consumer.Start()
	go barAggr.Start()

	<-ctx.Done()
}
