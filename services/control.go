package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/11me/calef/consumers/monitors"
	"github.com/11me/calef/manager"
	"github.com/11me/calef/models"
	"github.com/nats-io/nats.go"
)

type ControlService struct {
	nc      *nats.Conn
	manager *manager.Manager
	log     *slog.Logger
}

func NewControlService(ctx context.Context, nc *nats.Conn) *ControlService {
	return &ControlService{
		manager: manager.NewManager(ctx),
		nc:      nc,
		log:     slog.With("service", "ControlService"),
	}
}

func (svc *ControlService) SubmitPortfolio(ctx context.Context, portfolio *models.Portfolio) error {
	svc.log.Info(fmt.Sprintf("Submit portfolio (ID:%s)", portfolio.ID))

	m, err := monitors.NewPortfolioMonitor(ctx, svc.nc, portfolio)
	if err != nil {
		return fmt.Errorf("failed to create porfolio: %w", err)
	}

	err = svc.manager.Spawn(portfolio.ID, m)
	if err != nil {
		return fmt.Errorf("failed to spawn monitor: %w", err)
	}

	return nil
}

func (svc *ControlService) StopPortfolio(ctx context.Context, id string) error {
	svc.log.Info(fmt.Sprintf("Stop portfolio (ID:%s)", id))

	err := svc.manager.Evict(id)
	if err != nil {
		return fmt.Errorf("failed to evict portfolio with id %q: %w", id, err)
	}

	return nil
}
