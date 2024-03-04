package internal

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func getConn() *redis.Client {
	host := os.Getenv("REDIS_HOST")

	return redis.NewClient(&redis.Options{
		Addr:     host,
		Password: "",
		DB:       0,
	})
}

func GetUserDetail(ctx context.Context, token string) (*User, error) {
	conn := getConn()
	defer conn.Close()

	// Check and fetch if details exists in redis
	userByte, err := conn.Get(ctx, token).Bytes()
	if err == nil {
		user, err := ByteToUser(userByte)
		if err == nil {
			return user, nil
		}
	}

	// fetch user details via gRPC
	grpcResponse, err := fetchUserDetail(token)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:    grpcResponse.Id,
		Name:  grpcResponse.Name,
		Email: grpcResponse.Email,
	}

	userByte, err = UserToByte(*user)
	if err != nil {
		return nil, err
	}

	// Save user detail into redis
	if err := conn.Set(ctx, token, userByte, 1*time.Hour).Err(); err != nil {
		return nil, err
	}
	return user, nil
}
