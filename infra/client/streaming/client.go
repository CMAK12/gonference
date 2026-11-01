package streaming

import (
	"context"
	"fmt"

	"github.com/CMAK12/gonference/infra/client"
	gnpb "github.com/CMAK12/gonference/internal/gen/streaming/v1"

	"google.golang.org/grpc"
)

type Client struct {
	conn *grpc.ClientConn

	Signaling gnpb.SignalingClient
	Relay     gnpb.RelayClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := client.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("streaming.New: dial %s: %w", addr, err)
	}

	return &Client{
		conn:      conn,
		Signaling: gnpb.NewSignalingClient(conn),
		Relay:     gnpb.NewRelayClient(conn),
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

func (c *Client) ConnectRelay(ctx context.Context, req *gnpb.ConnectRelayRequest) (*gnpb.ConnectRelayResponse, error) {
	resp, err := c.Relay.Connect(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("streaming.ConnectRelay: %w", err)
	}

	if resp.Sdp == "" {
		return nil, fmt.Errorf("streaming.ConnectRelay: empty answer SDP")
	}

	return resp, nil
}

func (c *Client) DisconnectRelay(ctx context.Context, req *gnpb.DisconnectRelayRequest) error {
	if _, err := c.Relay.Disconnect(ctx, req); err != nil {
		return fmt.Errorf("streaming.DisconnectRelay: %w", err)
	}

	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
