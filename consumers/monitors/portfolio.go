package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	c "github.com/11me/calef/common"
	"github.com/11me/calef/consumers"
	"github.com/11me/calef/models"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/nats-io/nats.go"
)

type PortfolioMonitor struct {
	ctx             context.Context
	nc              *nats.Conn
	log             *slog.Logger
	portfolio       *models.Portfolio
	compiledProgram *vm.Program
	consumer        *consumers.Consumer

	// currentBars holds the current bar for each symbol.
	// NOTE: we don't need mutex here, because handler called sequentially for each symbol.
	currentBars map[string]*models.Bar
}

func NewPortfolioMonitor(ctx context.Context, nc *nats.Conn, portfilo *models.Portfolio) (*PortfolioMonitor, error) {
	program, err := expr.Compile(portfilo.Formula, expr.AllowUndefinedVariables())
	if err != nil {
		return nil, fmt.Errorf("failed to compile formula %q: %w", portfilo.Formula, err)
	}

	pm := &PortfolioMonitor{
		ctx:             ctx,
		nc:              nc,
		log:             slog.With("service", "PortfolioMonitor", "timeframe", portfilo.Timeframe.String()),
		portfolio:       portfilo,
		compiledProgram: program,
		currentBars:     make(map[string]*models.Bar),
		consumer:        consumers.NewConsumer(ctx, nc),
	}

	pm.consumer.
		SetLogger(pm.log).
		SetConcurrency(1)

	for _, symbol := range pm.portfolio.Symbols {
		pm.consumer.Subscribe(c.BinanceBarsSubj(symbol, pm.portfolio.Timeframe), pm)
	}

	return pm, nil
}

func (pm *PortfolioMonitor) Spawn() error {
	return pm.consumer.Start()
}

func (pm *PortfolioMonitor) Stop() error {
	return pm.consumer.Stop()
}

func (pm *PortfolioMonitor) Handle(msg *nats.Msg) error {
	var bar models.Bar
	if err := json.Unmarshal(msg.Data, &bar); err != nil {
		pm.log.Error("failed to unmarshal bar", "err", err, "data", string(msg.Data))
		return err
	}

	symbol := bar.Symbol

	// Update or create the aggregated bar for this symbol.
	currentBar, exists := pm.currentBars[symbol]
	if !exists || bar.StartTime.After(currentBar.StartTime) {
		// New bucket; replace the old bar.
		pm.currentBars[symbol] = &bar
	} else {
		// Same time bucket; aggregate the values.
		if bar.High > currentBar.High {
			currentBar.High = bar.High
		}

		if bar.Low < currentBar.Low {
			currentBar.Low = bar.Low
		}

		currentBar.Close = bar.Close
		currentBar.Volume += bar.Volume
	}

	// Only calculate the synthetic bar if bars for all symbols are available.
	if len(pm.currentBars) < len(pm.portfolio.Symbols) {
		pm.log.Debug("not all symbols have current bars", "current", len(pm.currentBars), "expected", len(pm.portfolio.Symbols))

		return nil
	}

	openParams := make(map[string]float64)
	highParams := make(map[string]float64)
	lowParams := make(map[string]float64)
	closeParams := make(map[string]float64)
	volumeSum := float64(0)

	for _, sym := range pm.portfolio.Symbols {
		b, ok := pm.currentBars[strings.ToLower(sym)]
		if !ok {
			continue
		}

		openParams[sym] = b.Open
		highParams[sym] = b.High
		lowParams[sym] = b.Low
		closeParams[sym] = b.Close
		volumeSum += b.Volume
	}

	syntheticOpen, err := pm.evalFormula(openParams)
	if err != nil {
		pm.log.Error("failed to evaluate formula for open", "err", err)
		return err
	}

	syntheticHigh, err := pm.evalFormula(highParams)
	if err != nil {
		pm.log.Error("failed to evaluate formula for high", "err", err)
		return err
	}

	syntheticLow, err := pm.evalFormula(lowParams)
	if err != nil {
		pm.log.Error("failed to evaluate formula for low", "err", err)
		return err
	}

	syntheticClose, err := pm.evalFormula(closeParams)
	if err != nil {
		pm.log.Error("failed to evaluate formula for close", "err", err)
		return err
	}

	syntheticBar := models.Bar{
		Symbol:    pm.portfolio.Formula + ".synth",
		Open:      syntheticOpen,
		High:      syntheticHigh,
		Low:       syntheticLow,
		Close:     syntheticClose,
		Volume:    volumeSum,
		StartTime: bar.StartTime,
		IsClosed:  false,
	}

	// TODO: Handle the bar somehow. Trigger notification or smth.
	subjSynthetic := fmt.Sprintf("synthetic.bars.%s", pm.portfolio.Timeframe.String())
	if err := pm.publishBar(subjSynthetic, &syntheticBar); err != nil {
		pm.log.Error("failed to publish synthetic bar", "subject", subjSynthetic, "err", err)
		return err
	}

	return nil
}

func (pm *PortfolioMonitor) evalFormula(parameters map[string]float64) (float64, error) {
	result, err := vm.Run(pm.compiledProgram, parameters)
	if err != nil {
		return 0, err
	}

	resFloat, ok := result.(float64)
	if !ok {
		return 0, fmt.Errorf("formula result is not a float64: %v", result)
	}

	return resFloat, nil
}

func (pm *PortfolioMonitor) publishBar(subj string, bar *models.Bar) error {
	data, err := json.Marshal(bar)
	if err != nil {
		return err
	}

	return pm.nc.Publish(subj, data)
}
