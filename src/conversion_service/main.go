package main

import (
	"bytes"
	"context"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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
	// Convert the media file
	key, err := convertMediaFile(in.GetKey(), in.GetIsAudioFile())
	if err != nil {
		log.Errorln("error caught while converting the media file: ", err)
		return nil, status.Errorf(codes.Internal, "something went wrong")
	}

	return &pb.ConversionResponse{Key: *key}, nil
}

func getS3Client() (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	// Create an Amazon S3 service client
	return s3.NewFromConfig(cfg), nil
}

func downloadFileFromS3(client *s3.Client, bucket, key, downloadPath string) error {
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

	return nil
}

func uploadFileToS3(client *s3.Client, bucket, key, filePath, ContentType string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	// Upload the given file to s3
	if _, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        file,
		ACL:         types.ObjectCannedACLPublicRead,
		ContentType: aws.String(ContentType),
	}); err != nil {
		return err
	}

	file.Close()

	return nil
}

func deleteFileFromS3(client *s3.Client, bucket, key string) {
	_, err := client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		log.Errorln("error caught while deleting file from s3: ", err, "key: ", key)
	}
}

func removeFiles(fileNames []string) {
	for _, fileName := range fileNames {
		err := os.Remove(fileName)
		if err != nil {
			log.Errorln("error caught while removing file: ", fileName)
		}
	}
}

func convertMediaFile(key string, isAudioFile bool) (*string, error) {
	client, err := getS3Client()
	if err != nil {
		return nil, err
	}

	bucket := os.Getenv("AWS_BUCKET_NAME")

	hlsKey := strings.Split(key, ".")[0] + ".m3u8"
	tsKey := strings.Split(key, ".")[0] + ".ts"

	srcFileName := strings.Split(key, "/")[1]
	dstFileName := strings.Split(hlsKey, "/")[1]
	tsFileName := strings.Split(tsKey, "/")[1]

	// Download the file from s3
	if err = downloadFileFromS3(client, bucket, key, srcFileName); err != nil {
		return nil, err
	}

	log.Infof("%s object downloaded successfully", key)

	var convertCmd *exec.Cmd
	var stderr bytes.Buffer

	// Determine conversion command based on file type
	if isAudioFile {
		// FFmpeg command for converting audio to AAC and then to HLS
		convertCmd = exec.Command("ffmpeg", "-i", srcFileName, "-c:a", "aac", "-b:a", "320k", "-vn", "-hls_time", "10", "-hls_playlist_type", "vod", "-hls_flags", "single_file", dstFileName)
	} else {
		// FFmpeg command for converting video to H.265/HEVC and then to HLS
		convertCmd = exec.Command("ffmpeg", "-i", srcFileName, "-c:v", "libx265", "-crf", "28", "-c:a", "aac", "-b:a", "320k", "-hls_time", "10", "-hls_playlist_type", "vod", "-hls_flags", "single_file", dstFileName)
	}

	convertCmd.Stderr = &stderr

	// Execute conversion
	if err = convertCmd.Run(); err != nil {
		log.Errorln("FFmpeg stderr: ", stderr.String())
		return nil, err
	}

	// Upload HLS file to s3
	if err = uploadFileToS3(client, bucket, hlsKey, dstFileName, "application/vnd.apple.mpegurl"); err != nil {
		return nil, err
	}

	log.Infof("%s object uploaded successfully", hlsKey)

	// Upload TS segment file to s3
	if err = uploadFileToS3(client, bucket, tsKey, tsFileName, "video/mp2t"); err != nil {
		return nil, err
	}

	log.Infof("%s object uploaded successfully", tsKey)

	// Remove old media file from s3
	go deleteFileFromS3(client, bucket, key)

	// Remove the downloaded or processed files in background
	go removeFiles([]string{srcFileName, dstFileName, tsFileName})

	return &hlsKey, err
}
