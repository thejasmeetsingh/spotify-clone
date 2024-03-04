package internal

import (
	"context"
	"flag"
	"os"

	"github.com/thejasmeetsingh/spotify-clone/src/user_service/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var addr = flag.String("addr", os.Getenv("GRPC_SERVER_ADDRESS"), "the gRPC server address")

func fetchUserDetail(token string) (*pb.UserDetailResponse, error) {
	flag.Parse()

	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	c := pb.NewUserServiceClient(conn)
	md := metadata.Pairs("authorization", "Bearer "+os.Getenv("GRPC_AUTH_KEY"))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	var header, trailer metadata.MD
	r, err := c.UserDetail(ctx, &pb.UserDetailRequest{Token: token}, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		return nil, err
	}

	return r, nil
}
