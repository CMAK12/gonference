// Package streaming holds the gateway's client for the streaming service's
// gRPC signaling API. The gateway uses it to forward WHIP offers to the SFU.
package streaming

import (
	"context"
	"fmt"

	impb "github.com/CMAK12/gonference/internal/gen/streaming/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client is a gRPC client to the streaming service's Signaling API.
type Client struct {
	conn      *grpc.ClientConn
	signaling impb.SignalingClient
}

// New dials the streaming service at addr (host:port) and returns a Client.
func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("streaming.New: dial %s: %w", addr, err)
	}

	return &Client{
		conn:      conn,
		signaling: impb.NewSignalingClient(conn),
	}, nil
}

// Connect forwards a publisher's SDP offer to the SFU and returns the SFU's
// answer SDP. Post-handshake signaling continues over the peer's DataChannel.
func (c *Client) Connect(ctx context.Context, roomID, memberID, offer string) (string, error) {
	resp, err := c.signaling.Connect(ctx, &impb.SignalMessage{
		Type:     impb.MessageType_OFFER,
		RoomId:   roomID,
		MemberId: memberID,
		Sdp:      &offer,
	})
	if err != nil {
		return "", fmt.Errorf("streaming.Connect: %w", err)
	}

	if resp.Sdp == nil {
		return "", fmt.Errorf("streaming.Connect: empty answer SDP")
	}

	return *resp.Sdp, nil
}

// Close tears down the underlying gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}
