package ulasan

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	publicReviewsVersionKey = "ulasan:public:version"
	publicReviewsCacheTTL   = time.Minute
)

var ErrPublicCacheMiss = errors.New("public review cache miss")

type PublicReviewsCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, expiration time.Duration) error
	Increment(ctx context.Context, key string) (int64, error)
}

type RedisPublicReviewsCache struct {
	client *redis.Client
}

func NewRedisPublicReviewsCacheFromEnv() (*RedisPublicReviewsCache, error) {
	rawURL := strings.TrimSpace(os.Getenv("REDIS_URL"))
	if rawURL == "" {
		return nil, nil
	}

	options, err := redis.ParseURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("REDIS_URL tidak valid: %w", err)
	}

	client := redis.NewClient(options)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("gagal terhubung ke Redis: %w", err)
	}

	return &RedisPublicReviewsCache{client: client}, nil
}

func (c *RedisPublicReviewsCache) Get(ctx context.Context, key string) (string, error) {
	value, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrPublicCacheMiss
	}
	return value, err
}

func (c *RedisPublicReviewsCache) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *RedisPublicReviewsCache) Increment(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

func (c *RedisPublicReviewsCache) Close() error {
	return c.client.Close()
}

func publicReviewsCacheKey(version string, page, limit, rating int) string {
	return fmt.Sprintf("ulasan:public:v%s:p%d:l%d:r%d", version, page, limit, rating)
}
