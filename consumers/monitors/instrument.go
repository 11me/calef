package monitors

import (
	"context"

	"github.com/11me/calef/models"
	"github.com/nats-io/nats.go"
)

type InstrumentMonitor struct {
	ctx    context.Context
	symbol string
	tf     models.Timeframe
}

func NewInstrumentMonitor(ctx context.Context, symbol string, tf models.Timeframe) *InstrumentMonitor {
	return &InstrumentMonitor{
		ctx:    ctx,
		symbol: symbol,
		tf:     tf,
	}
}

func (m *InstrumentMonitor) Handle(msg *nats.Msg) error {
	return nil
}
