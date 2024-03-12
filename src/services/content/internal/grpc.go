package internal

import (
	"context"
	"flag"
	"os"

	conversionPB "github.com/thejasmeetsingh/spotify-clone/src/services/conversion/pb"
	userPB "github.com/thejasmeetsingh/spotify-clone/src/services/user/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var userAddr = flag.String("userAddr", os.Getenv("USER_GRPC_ADDRESS"), "user gRPC server address")
var conversionAddr = flag.String("conversionAddr", os.Getenv("CONVERSION_GRPC_ADDRESS"), "conversion gRPC server address")

// gRPC to user service to fetch user details
func fetchUserDetail(token string) (*userPB.UserDetailResponse, error) {
	flag.Parse()

	conn, err := grpc.Dial(*userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	c := userPB.NewUserServiceClient(conn)
	md := metadata.Pairs("authorization", "Bearer "+os.Getenv("GRPC_AUTH_KEY"))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	var header, trailer metadata.MD
	r, err := c.UserDetail(ctx, &userPB.UserDetailRequest{Token: token}, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		return nil, err
	}

	return r, nil
}

// gRPC to conversion service and process the uploaded file
func processContentMedia(key string, isAudioFile bool) (*conversionPB.ConversionResponse, error) {
	flag.Parse()

	conn, err := grpc.Dial(*conversionAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	c := conversionPB.NewConversionServiceClient(conn)
	md := metadata.Pairs("authorization", "Bearer "+os.Getenv("GRPC_AUTH_KEY"))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	var header, trailer metadata.MD
	r, err := c.Conversion(ctx, &conversionPB.ConversionRequest{Key: key, IsAudioFile: isAudioFile}, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		return nil, err
	}

	return r, nil
}
