package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"tiny-bitly/internal/cache"
	"tiny-bitly/internal/dao"
	"tiny-bitly/internal/model"

	"github.com/redis/go-redis/v9"
)

// URLRecordCachedDAO wraps a URLRecordDAO with Redis caching for read operations.
type URLRecordCachedDAO struct {
	underlying dao.URLRecordDAO
	redis      *redis.Client
}

// NewURLRecordCachedDAO creates a new cached DAO that wraps the underlying DAO.
func NewURLRecordCachedDAO(underlying dao.URLRecordDAO) (*URLRecordCachedDAO, error) {
	redisClient := cache.GetClient()
	if redisClient == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}

	return &URLRecordCachedDAO{
		underlying: underlying,
		redis:      redisClient,
	}, nil
}

// Create delegates to the underlying DAO and invalidates the cache for the short code.
func (d *URLRecordCachedDAO) Create(ctx context.Context, urlRecord model.URLRecord) (*model.URLRecordEntity, error) {
	entity, err := d.underlying.Create(ctx, urlRecord)
	if err != nil {
		return nil, err
	}

	// Cache the newly created record
	if entity != nil {
		if err := d.setCache(ctx, entity); err != nil {
			// Log but don't fail the create operation if caching fails
			slog.Warn("Failed to cache created record", "error", err, "shortCode", urlRecord.ShortCode)
		}
	}

	return entity, nil
}

// GetByShortCode checks Redis cache first, then falls back to the underlying DAO.
// Note: We don't use mutexes here - allowing some duplicate DB queries under
// extreme load is better than serializing all requests. Redis handles concurrent
// reads efficiently, and the database connection pool handles concurrent queries.
func (d *URLRecordCachedDAO) GetByShortCode(ctx context.Context, shortCode string) (*model.URLRecordEntity, error) {
	// Try to get from cache first
	cachedEntity, err := d.getFromCache(ctx, shortCode)
	if err == nil && cachedEntity != nil {
		// Cache hit - verify it hasn't expired
		if !cachedEntity.IsExpired() {
			return cachedEntity, nil
		}
		// Expired - remove from cache and fall through to database
		d.deleteFromCache(ctx, shortCode)
	} else if err != redis.Nil {
		// Redis error (not a cache miss) - log but continue to database
		// This could indicate Redis is down or overloaded
		slog.Debug("Redis error during cache lookup", "error", err, "shortCode", shortCode)
	}

	// Cache miss or expired - get from database
	// Multiple concurrent requests for the same key may all hit the DB,
	// but that's acceptable - the DB connection pool handles it efficiently.
	entity, err := d.underlying.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	// Cache the result if found
	if entity != nil {
		if err := d.setCache(ctx, entity); err != nil {
			// Log but don't fail the read operation if caching fails
			slog.Debug("Failed to cache retrieved record", "error", err, "shortCode", shortCode)
		}
	}

	return entity, nil
}

// getCacheKey returns the Redis key for a short code.
func (d *URLRecordCachedDAO) getCacheKey(shortCode string) string {
	return fmt.Sprintf("url:%s", shortCode)
}

// getFromCache retrieves a URL record from Redis.
func (d *URLRecordCachedDAO) getFromCache(ctx context.Context, shortCode string) (*model.URLRecordEntity, error) {
	key := d.getCacheKey(shortCode)
	val, err := d.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var entity model.URLRecordEntity
	if err := json.Unmarshal([]byte(val), &entity); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached record: %w", err)
	}

	return &entity, nil
}

// setCache stores a URL record in Redis with TTL based on expiration time.
func (d *URLRecordCachedDAO) setCache(ctx context.Context, entity *model.URLRecordEntity) error {
	key := d.getCacheKey(entity.ShortCode)

	// Serialize the entity
	data, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("failed to marshal entity for cache: %w", err)
	}

	// Calculate TTL - use the time until expires_at, with a minimum of 1 second
	now := time.Now()
	ttl := entity.ExpiresAt.Sub(now)
	if ttl <= 0 {
		// Already expired, don't cache
		return nil
	}
	// Add a small buffer (1 second) to ensure we don't serve expired records
	ttl = ttl + time.Second

	// Set in Redis with TTL
	if err := d.redis.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// deleteFromCache removes a URL record from Redis.
func (d *URLRecordCachedDAO) deleteFromCache(ctx context.Context, shortCode string) {
	key := d.getCacheKey(shortCode)
	if err := d.redis.Del(ctx, key).Err(); err != nil {
		slog.Warn("Failed to delete from cache", "error", err, "shortCode", shortCode)
	}
}
