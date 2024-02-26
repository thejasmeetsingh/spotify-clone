package main

import (
	"net"
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/thejasmeetsingh/spotify-clone/src/user_service/pb"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedUserServiceServer
}

func main() {
	godotenv.Load()

	grpc_port := os.Getenv("GRPC_PORT")

	if grpc_port == "" {
		log.Fatalln("gRPC port is not configured")
	}

	lis, err := net.Listen("tcp", ":"+grpc_port)
	if err != nil {
		log.Fatalln("failed to listen gRPC: ", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, &server{})

	log.Infoln("gRPC service is up & running")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalln("failed to serve gRPC: ", err)
	}
}
