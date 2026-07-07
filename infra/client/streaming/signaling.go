package streaming

import (
	"context"
	"fmt"

	impb "github.com/CMAK12/gonference/internal/gen/streaming/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn      *grpc.ClientConn
	signaling impb.SignalingClient
}

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
