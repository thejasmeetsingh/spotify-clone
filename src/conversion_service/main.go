package main

import (
	"context"
	"net"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/thejasmeetsingh/spotify-clone/src/conversion_service/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedConversionServiceServer
}

func valid(authorization []string) bool {
	if len(authorization) < 1 {
		return false
	}
	token := strings.TrimPrefix(authorization[0], "Bearer ")
	return token == os.Getenv("GRPC_AUTH_KEY")
}

func ensureValidToken(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
	}
	if !valid(md["authorization"]) {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}
	// Continue execution of handler after ensuring a valid token.
	return handler(ctx, req)
}

func main() {
	lis, err := net.Listen("tcp", ":"+os.Getenv("GRPC_PORT"))
	if err != nil {
		log.Fatalln("failed to listen gRPC: ", err)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(ensureValidToken),
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterConversionServiceServer(grpcServer, &server{})

	log.Infoln("gRPC service is up & running")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalln("failed to serve gRPC: ", err)
	}
}
