package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

// gRPC request handler
func (s *server) Conversion(ctx context.Context, in *pb.ConversionRequest) (*pb.ConversionResponse, error) {
	// Convert the media file in background
	go convertMediaFile(in.GetKey(), in.GetIsAudioFile())

	return &pb.ConversionResponse{
		Message: "Request received successfully, Processing media file",
	}, nil
}

func getS3Client() (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return nil, err
	}
	// Create an Amazon S3 service client
	return s3.NewFromConfig(cfg), nil
}

func downloadFile(client *s3.Client, bucket, key, downloadPath string) error {
	// Fetch the file from s3
	result, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	defer result.Body.Close()

	// Create a file to write the download to
	file, err := os.Create(downloadPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the contents of S3 Object to the file
	if _, err = file.ReadFrom(result.Body); err != nil {
		return err
	}

	log.Infof("%s object downloaded successfully", key)
	return nil
}

func uploadFile(client *s3.Client, bucket, key, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalln("error while opening the given file: ", err)
		return
	}

	// Upload the given file to s3
	if _, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	}); err != nil {
		log.Fatalln("error caught while uploading file to s3: ", err)
		return
	}

	log.Infof("%s object uploaded successfully", key)
	file.Close()

	if err = os.Remove(filePath); err != nil {
		log.Errorln("error caught while deleting the file", err)
	}
}

func convertMediaFile(key string, isAudioFile bool) {
	client, err := getS3Client()
	if err != nil {
		log.Fatalln("error caught while creating s3 client: ", err)
		return
	}

	bucket := os.Getenv("AWS_BUCKET")
	downloadPath := fmt.Sprintf("/media/%s", key)
	outputPath := fmt.Sprintf("/output/%s", strings.Split(key, ".")[0]+".m3u8")

	// Download the file from s3
	downloadFile(client, bucket, key, downloadPath)

	var convertCmd *exec.Cmd

	// Determine conversion command based on file type
	if isAudioFile {
		// FFmpeg command for converting audio to AAC and then to HLS
		convertCmd = exec.Command("ffmpeg", "-i", downloadPath, "-c:a", "aac", "-b:a", "320k", "-vn", "-hls_time", "10", "-hls_playlist_type", "vod", outputPath)
	} else {
		// FFmpeg command for converting video to H.265/HEVC and then to HLS
		convertCmd = exec.Command("ffmpeg", "-i", downloadPath, "-c:v", "libx265", "-crf", "28", "-c:a", "aac", "-b:a", "320k", "-hls_time", "10", "-hls_playlist_type", "vod", outputPath)
	}

	// Execute conversion
	if err := convertCmd.Run(); err != nil {
		log.Fatalln("error caught while converting the media file: ", err)
		return
	}

	// Upload file back to s3 in background
	go uploadFile(client, bucket, key, outputPath)
}
