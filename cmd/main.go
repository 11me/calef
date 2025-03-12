package main

import (
	"context"
	"log"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/11me/calef/config"
	"github.com/11me/calef/consumers/aggregators"
	"github.com/11me/calef/consumers/exchange"
	"github.com/11me/calef/consumers/monitors"
	"github.com/11me/calef/models"
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

	binanceConsumer := exchange.NewBinanceConsumer(ctx, nc)
	binanceConsumer.SubscribeTicks("btcusdt", "ethusdt")

	btcAgg := aggregators.NewBarAggregator(ctx, nc, "btcusdt", models.M1)
	ethAgg := aggregators.NewBarAggregator(ctx, nc, "ethusdt", models.M1)

	ethBtcPortfolio, err := monitors.NewPortfolioMonitor(ctx, nc, []string{"btcusdt", "ethusdt"}, models.M1, "ethusdt/btcusdt")
	if err != nil {
		log.Fatal(err)
	}

	go binanceConsumer.Start()
	go btcAgg.Start()
	go ethAgg.Start()
	go ethBtcPortfolio.Start()

	<-ctx.Done()
}
