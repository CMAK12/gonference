package streaming

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/CMAK12/gonference/infra/client"
	gnpb "github.com/CMAK12/gonference/internal/gen/streaming/v1"

	"google.golang.org/grpc"
)

type Client struct {
	conn *grpc.ClientConn

	Signaling gnpb.SignalingClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := client.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("streaming.New: dial %s: %w", addr, err)
	}

	return &Client{
		conn:      conn,
		Signaling: gnpb.NewSignalingClient(conn),
	}, nil
}

func (c *Client) Connect(ctx context.Context, req *gnpb.SignalMessage) (*gnpb.SignalMessage, error) {
	resp, err := c.Signaling.Connect(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("streaming.Connect: %w", err)
	}

	if resp.Sdp == nil {
		return nil, fmt.Errorf("streaming.Connect: empty answer SDP")
	}

	return resp, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

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
