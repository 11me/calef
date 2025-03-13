package manager

import (
	"context"
	"errors"
	"sync"
)

type Spawnable interface {
	Spawn() error
	Stop() error
}

type Manager struct {
	ctx   context.Context
	mu    sync.RWMutex
	items map[string]Spawnable
}

func NewManager(ctx context.Context) *Manager {
	return &Manager{
		ctx:   ctx,
		items: make(map[string]Spawnable),
	}
}

func (s *Manager) Spawn(id string, item Spawnable) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.items[id]; exists {
		return errors.New("spawnable with id already exists: " + id)
	}

	if err := item.Spawn(); err != nil {
		return err
	}

	s.items[id] = item

	return nil
}

func (s *Manager) Evict(id string) error {
	s.mu.Lock()
	item, exists := s.items[id]
	if !exists {
		return nil
	}
	delete(s.items, id)
	s.mu.Unlock()

	if err := item.Stop(); err != nil {
		return err
	}

	return nil
}

func (s *Manager) StopAll() error {
	s.mu.Lock()
	items := s.items
	s.items = make(map[string]Spawnable)
	s.mu.Unlock()

	var ee error
	for _, item := range items {
		if err := item.Stop(); err != nil {
			ee = errors.Join(ee, err)
		}
	}

	return ee
}
