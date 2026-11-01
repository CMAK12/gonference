package conference

import (
	"context"
	"fmt"

	"github.com/CMAK12/gonference/infra/client"
	pb "github.com/CMAK12/gonference/internal/gen/conference/v1"
	"google.golang.org/grpc"
)

type Client struct {
	conn *grpc.ClientConn

	Conference pb.ConferenceClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := client.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("conference.New: dial %s: %w", addr, err)
	}

	return &Client{
		conn:       conn,
		Conference: pb.NewConferenceClient(conn),
	}, nil
}

func (c *Client) CreateConference(ctx context.Context, req *pb.CreateConferenceRequest) (*pb.CreateConferenceResponse, error) {
	resp, err := c.Conference.CreateConference(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("conference.CreateConference: %w", err)
	}

	return resp, nil
}

func (c *Client) JoinConference(ctx context.Context, req *pb.JoinConferenceRequest) (*pb.JoinConferenceResponse, error) {
	resp, err := c.Conference.JoinConference(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("conference.JoinConference: %w", err)
	}

	return resp, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
