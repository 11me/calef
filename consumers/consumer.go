package consumers

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/nats-io/nats.go"
)

type Handler interface {
	Handle(*nats.Msg) error
}

type HandlerFunc func(*nats.Msg) error

func (h HandlerFunc) Handle(msg *nats.Msg) error { return h(msg) }

type handlerEntry struct {
	subj    string
	handler Handler
}

type Consumer struct {
	ctx          context.Context
	nc           *nats.Conn
	handlers     map[string]*handlerEntry
	waitHandlers sync.WaitGroup
	log          *slog.Logger
	subs         []*nats.Subscription
	sem          chan struct{}
	done         chan struct{}
}

func NewConsumer(ctx context.Context, nc *nats.Conn) *Consumer {
	return &Consumer{
		ctx:      ctx,
		nc:       nc,
		done:     make(chan struct{}),
		log:      slog.With("service", "consumer"),
		handlers: make(map[string]*handlerEntry),
	}
}

// Subscribe registers the handler for all its subjects with an optional concurrency limit.
func (c *Consumer) Subscribe(subject string, h Handler) *Consumer {
	entry := &handlerEntry{
		subj:    subject,
		handler: h,
	}

	c.handlers[subject] = entry

	return c
}

func (c *Consumer) SetLogger(log *slog.Logger) *Consumer { c.log = log; return c }

func (c *Consumer) SetConcurrency(concrrency uint32) *Consumer {
	if concrrency > 0 {
		c.sem = make(chan struct{}, concrrency)
	}

	return c
}

func (c *Consumer) Start() error {
	for subj, entry := range c.handlers {
		e := entry
		sub, err := c.nc.Subscribe(subj, func(msg *nats.Msg) {
			if c.sem != nil {
				c.sem <- struct{}{}
				defer func() { <-c.sem }()
			}

			select {
			case <-c.done:
				c.log.Debug(fmt.Sprintf("shutting down consumer on %s", e.subj))
				return
			default:
			}

			c.waitHandlers.Add(1)
			defer c.waitHandlers.Done()

			if err := e.handler.Handle(msg); err != nil {
				c.log.Error("failed to handle message", "subject", e.subj, "err", err)
				return
			}
		})
		if err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", subj, err)
		}

		c.subs = append(c.subs, sub)
	}

	// Wait for cancellation.
	<-c.ctx.Done()
	close(c.done)
	c.waitHandlers.Wait()

	var drainErr error
	for _, sub := range c.subs {
		if err := sub.Drain(); err != nil {
			c.log.Error("failed to drain subscription", "err", err)
			drainErr = err
		}
	}

	return drainErr
}
