package redis

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis"
)

var (
	client *redis.Client
)

func InitRedis() error {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		return nil
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		log.Fatal("REDIS_PORT env variable is empty")
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisScheme := os.Getenv("REDIS_SCHEME")
	if redisScheme == "" {
		redisScheme = "rediss"
	}

	connStr := fmt.Sprintf("%s://default:%s@%s:%s", redisScheme, redisPassword, redisHost, redisPort)
	opts, err := redis.ParseURL(connStr)
	if err != nil {
		return err
	}

	client = redis.NewClient(opts)
	_, err = client.Ping().Result()

	return err
}

func Set(key string, value interface{}) error {
	return client.Set(key, value, time.Duration(0)).Err()
}
