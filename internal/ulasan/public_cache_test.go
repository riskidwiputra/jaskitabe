package ulasan

import (
	"context"
	"testing"
	"time"
)

type fakePublicReviewsCache struct {
	incrementedKey string
	incremented    int
}

func (c *fakePublicReviewsCache) Get(context.Context, string) (string, error) {
	return "", ErrPublicCacheMiss
}

func (c *fakePublicReviewsCache) Set(context.Context, string, string, time.Duration) error {
	return nil
}

func (c *fakePublicReviewsCache) Increment(_ context.Context, key string) (int64, error) {
	c.incrementedKey = key
	c.incremented++
	return int64(c.incremented), nil
}

func TestInvalidatePublicReviewsIncrementsVersion(t *testing.T) {
	cache := &fakePublicReviewsCache{}
	handler := &Handler{publicCache: cache}

	handler.invalidatePublicReviews(context.Background())

	if cache.incremented != 1 {
		t.Fatalf("cache increment calls = %d, want 1", cache.incremented)
	}
	if cache.incrementedKey != publicReviewsVersionKey {
		t.Fatalf("incremented key = %q, want %q", cache.incrementedKey, publicReviewsVersionKey)
	}
}

func TestPublicReviewsCacheKeyIncludesQueryAndVersion(t *testing.T) {
	base := publicReviewsCacheKey("4", 1, 9, 0)
	tests := []string{
		publicReviewsCacheKey("5", 1, 9, 0),
		publicReviewsCacheKey("4", 2, 9, 0),
		publicReviewsCacheKey("4", 1, 12, 0),
		publicReviewsCacheKey("4", 1, 9, 5),
	}

	for _, key := range tests {
		if key == base {
			t.Fatalf("cache key %q should differ from %q", key, base)
		}
	}
}
