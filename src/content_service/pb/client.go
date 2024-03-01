package pb

import (
	"context"
	"flag"
	"os"

	"github.com/thejasmeetsingh/spotify-clone/src/user_service/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var addr = flag.String("addr", "localhost:8080", "the gRPC server address")

func getGrpcConn() (*grpc.ClientConn, error) {
	flag.Parse()
	return grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func GetUserDetail() (*pb.UserDetailResponse, error) {
	conn, err := getGrpcConn()
	if err != nil {
		return nil, err
	}

	c := pb.NewUserServiceClient(conn)
	md := metadata.Pairs("authorization", "Bearer "+os.Getenv("GRPC_AUTH_KEY"))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	var header, trailer metadata.MD
	r, err := c.UserDetail(ctx, &pb.UserDetailRequest{Token: ""}, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		return nil, err
	}

	return r, nil
}
