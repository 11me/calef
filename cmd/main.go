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
	"github.com/11me/calef/models"
	"github.com/11me/calef/server"
	"github.com/11me/calef/server/handlers"
	"github.com/11me/calef/services"
	"github.com/nats-io/nats.go"
)

var (
	timeframes = []models.Timeframe{models.M1, models.M5}
	symbols    = []string{
		"btcusdt", "ethusdt",
	}
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

	controlSvc := services.NewControlService(ctx, nc)

	binanceConsumer := exchange.NewBinanceConsumer(ctx, nc)
	binanceConsumer.SubscribeTicks(symbols...)

	srv := server.NewServer(conf.Addr)
	srv.HandleFunc("POST /api/portfolios", handlers.HandleSubmitPortfolio(controlSvc))
	srv.HandleFunc("DELETE /api/portfolios/{id}", handlers.HandleStopPortfolio(controlSvc))

	aggs := make([]*aggregators.BarAggregator, 0, len(symbols)*len(timeframes))
	for _, tf := range timeframes {
		for _, symbol := range symbols {
			agg := aggregators.NewBarAggregator(ctx, nc, symbol, tf)
			aggs = append(aggs, agg)
			agg.Spawn()
		}
	}

	go binanceConsumer.Start()
	go srv.Start()

	<-ctx.Done()

	for i := range aggs {
		aggs[i].Stop()
	}
}
