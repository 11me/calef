package main

import (
	"context"
	"log"
	"log/slog"
	"os/signal"
	"syscall"

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
	consumer.SubscribeTicks("btcusdt")
	consumer.SubscribeTicks("ethusdt")

	go consumer.Start()

	<-ctx.Done()
}
