package main

import (
	"context"
	"net"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
	"github.com/thejasmeetsingh/spotify-clone/src/user_service/database"
	"github.com/thejasmeetsingh/spotify-clone/src/user_service/pb"
	"github.com/thejasmeetsingh/spotify-clone/src/user_service/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedUserServiceServer
}

func valid(authorization []string) bool {
	if len(authorization) < 1 {
		return false
	}
	token := strings.TrimPrefix(authorization[0], "Bearer ")
	return token == "secret-auth-key"
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
	pb.RegisterUserServiceServer(grpcServer, &server{})

	log.Infoln("gRPC service is up & running")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalln("failed to serve gRPC: ", err)
	}
}

func dbConn(ctx context.Context) (*pgx.Conn, error) {
	return pgx.Connect(context.Background(), os.Getenv("DB_URL"))
}

func (s *server) UserDetail(ctx context.Context, in *pb.UserDetailRequest) (*pb.UserDetailResponse, error) {
	conn, err := dbConn(ctx)
	if err != nil {
		log.Fatalln("error while connecting to DB: ", err)
		return nil, status.Errorf(codes.Internal, "something went wrong")
	}
	defer conn.Close(ctx)

	dbCfg := &database.Config{
		DB:      conn,
		Queries: database.New(conn),
	}

	token := in.GetToken()

	// Verify the token and get the encoded payload which is the userID string
	claims, err := utils.VerifyToken(token)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "invalid token")
	}

	// Check the validity of the token
	if !time.Unix(claims.ExpiresAt.Unix(), 0).After(time.Now()) {
		return nil, status.Errorf(codes.PermissionDenied, "token is expired")
	}

	// Convert the userID string to UUID
	userID, err := uuid.Parse(claims.Data)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "invalid token")
	}

	// Fetch user by ID from DB
	dbUser, err := database.GetUserByIDFromDB(dbCfg, ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "something went wrong")
	}

	return &pb.UserDetailResponse{
		Id:    dbUser.ID.String(),
		Name:  dbUser.Name.String,
		Email: dbUser.Email,
	}, nil
}
