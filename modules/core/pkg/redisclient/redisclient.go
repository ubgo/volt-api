package redisclient

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func New(url string, db int) (*redis.Client, error) {
	// url := "redis://user:password@localhost:6379/0?protocol=3"
	opts, err := redis.ParseURL(url)
	if err != nil {

		return nil, err
	}

	opts.DB = db

	client := redis.NewClient(opts)

	return client, nil
}

// pingRedis sends a PING command to the Redis server to verify the connection.
func PingRedis(client *redis.Client) error {
	// The Ping() method returns a StatusCmd.
	// We call .Err() on the result to check for any errors.
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return fmt.Errorf("PING command failed: %w", err)
	}
	return nil
}
