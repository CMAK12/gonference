package streaming

import (
	"context"
	"errors"
	"fmt"
	"sync"

	gnpb "github.com/CMAK12/gonference/internal/gen/streaming/v1"
)

type Pool struct {
	mu      sync.Mutex
	clients map[string]*Client
}

func NewPool() *Pool {
	return &Pool{
		clients: make(map[string]*Client),
	}
}

func (p *Pool) Get(podAddr string) (*Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if conn, ok := p.clients[podAddr]; ok {
		return conn, nil
	}

	conn, err := NewClient(podAddr)
	if err != nil {
		return nil, err
	}
	p.clients[podAddr] = conn

	return conn, nil
}

func (p *Pool) Connect(ctx context.Context, podAddr string, req *gnpb.SignalMessage) (*gnpb.SignalMessage, error) {
	conn, err := p.Get(podAddr)
	if err != nil {
		return nil, fmt.Errorf("streaming.Connect: %w", err)
	}

	return conn.Connect(ctx, req)
}

func (p *Pool) ConnectRelay(ctx context.Context, podAddr string, req *gnpb.ConnectRelayRequest) (*gnpb.ConnectRelayResponse, error) {
	conn, err := p.Get(podAddr)
	if err != nil {
		return nil, fmt.Errorf("streaming.ConnectRelay: %w", err)
	}

	return conn.ConnectRelay(ctx, req)
}

func (p *Pool) DisconnectRelay(ctx context.Context, podAddr string, req *gnpb.DisconnectRelayRequest) error {
	conn, err := p.Get(podAddr)
	if err != nil {
		return fmt.Errorf("streaming.DisconnectRelay: %w", err)
	}

	return conn.DisconnectRelay(ctx, req)
}

func (p *Pool) Remove(podAddr string) error {
	p.mu.Lock()
	conn, ok := p.clients[podAddr]
	delete(p.clients, podAddr)
	p.mu.Unlock()

	if !ok {
		return nil
	}

	return conn.Close()
}

func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	errs := make([]error, 0, len(p.clients))
	for addr, conn := range p.clients {
		errs = append(errs, conn.Close())
		delete(p.clients, addr)
	}

	return errors.Join(errs...)
}
