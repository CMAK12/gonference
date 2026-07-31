package streaming

import (
	"context"
	"fmt"

	"github.com/CMAK12/gonference/infra/client"
	pb "github.com/CMAK12/gonference/internal/gen/streaming/v1"

	"google.golang.org/grpc"
)

type Client struct {
	conn *grpc.ClientConn

	Signaling pb.SignalingClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := client.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("streaming.New: dial %s: %w", addr, err)
	}

	return &Client{
		conn:      conn,
		Signaling: pb.NewSignalingClient(conn),
	}, nil
}

func (c *Client) Connect(ctx context.Context, roomID, memberID, offer string) (string, error) {
	resp, err := c.Signaling.Connect(ctx, &pb.SignalMessage{
		Type:     pb.MessageType_OFFER,
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

func (c *Client) Close() error {
	return c.conn.Close()
}
