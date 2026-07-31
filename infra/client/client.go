package client

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Dial(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	defaults := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),

		grpc.WithDefaultServiceConfig(`{
            "loadBalancingConfig": [
                {
                    "round_robin": {}
                }
            ]
        }`),
	}

	defaults = append(defaults, opts...)

	return grpc.NewClient("dns:///"+addr, defaults...)
}
