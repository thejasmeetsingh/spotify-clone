package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	log "github.com/sirupsen/logrus"
)

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

func uploadFile(client *s3.Client, bucket, key, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Upload the given file to s3
	if _, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	}); err != nil {
		return nil
	}

	log.Infof("%s object uploaded successfully", key)
	return nil
}

func getFileName(key string) string {
	filename := strings.Split(key, ".")[0]
	return filename + ".m3u8"
}

func ConvertMediaFile(key string, isAudioFile bool) {
	client, err := getS3Client()
	if err != nil {
		log.Fatalln("error caught while creating s3 client: ", err)
		return
	}

	bucket := os.Getenv("AWS_BUCKET")
	downloadPath := fmt.Sprintf("/media/%s", key)
	outputPath := fmt.Sprintf("/output/%s", getFileName(key))

	// Download the file from s3
	downloadFile(client, bucket, key, downloadPath)

	var convertCmd *exec.Cmd

	// Determine conversion command based on file type
	if isAudioFile {
		// FFmpeg command for converting audio to AAC
		convertCmd = exec.Command("ffmpeg", "-i", downloadPath, "-c:a", "aac", "-b:a", "320k", "-hls_time", "10", "-hls_playlist_type", outputPath)
	} else {
		// FFmpeg command for converting video to H.265/HEVC and then to HLS
		convertCmd = exec.Command("ffmpeg", "-i", downloadPath, "-c:v", "libx265", "-c:a", "aac", "-hls_time", "10", "-hls_playlist_type", "vod", outputPath)
	}

	// Execute conversion
	if err := convertCmd.Run(); err != nil {
		log.Fatalln("error caught while converting the media file: ", err)
		return
	}

	// Upload back to s3
	go uploadFile(client, bucket, key, downloadPath)
}
