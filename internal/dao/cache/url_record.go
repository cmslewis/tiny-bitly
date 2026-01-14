package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	redisCache "tiny-bitly/internal/cache"
	"tiny-bitly/internal/dao"
	"tiny-bitly/internal/model"

	"github.com/redis/go-redis/v9"
)

// URLRecordCachedDAO wraps a URLRecordDAO with Redis caching for read operations.
type URLRecordCachedDAO struct {
	underlying     dao.URLRecordDAO
	redis          *redis.Client
	circuitBreaker *redisCache.CircuitBreaker
}

// NewURLRecordCachedDAO creates a new cached DAO that wraps the underlying DAO.
func NewURLRecordCachedDAO(underlying dao.URLRecordDAO) (*URLRecordCachedDAO, error) {
	redisClient := redisCache.GetClient()
	if redisClient == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}

	return &URLRecordCachedDAO{
		underlying:     underlying,
		redis:          redisClient,
		circuitBreaker: redisCache.NewCircuitBreaker(),
	}, nil
}

// Create delegates to the underlying DAO and invalidates the cache for the short code.
func (d *URLRecordCachedDAO) Create(ctx context.Context, urlRecord model.URLRecord) (*model.URLRecordEntity, error) {
	entity, err := d.underlying.Create(ctx, urlRecord)
	if err != nil {
		return nil, err
	}

	// Cache the newly created record asynchronously (non-blocking - failures don't affect create).
	// Fire-and-forget to avoid blocking the write response.
	if entity != nil && !d.circuitBreaker.IsOpen() {
		go func() {
			// Use background context since this is async.
			if err := d.setCache(context.Background(), entity); err != nil {
				d.circuitBreaker.RecordFailure()
				slog.Warn("Failed to cache created record", "error", err, "shortCode", urlRecord.ShortCode, "circuitState", d.circuitBreaker.GetState())
			} else {
				d.circuitBreaker.RecordSuccess()
			}
		}()
	}

	return entity, nil
}

// GetByShortCode checks Redis cache first, then falls back to the underlying DAO.
// Uses circuit breaker to prevent cascading failures when Redis is down.
func (d *URLRecordCachedDAO) GetByShortCode(ctx context.Context, shortCode string) (*model.URLRecordEntity, error) {
	// Check circuit breaker - if open, skip Redis and go straight to DB
	if !d.circuitBreaker.IsOpen() {
		// Try to get from cache first
		cachedEntity, err := d.getFromCache(ctx, shortCode)
		if err == nil && cachedEntity != nil {
			// Cache hit - verify it hasn't expired
			if !cachedEntity.IsExpired() {
				d.circuitBreaker.RecordSuccess()
				return cachedEntity, nil
			}
			// Expired - remove from cache and fall through to database
			d.deleteFromCache(ctx, shortCode)
			d.circuitBreaker.RecordSuccess() // Cache operation succeeded
		} else if err != redis.Nil {
			// Redis error (not a cache miss) - record failure
			d.circuitBreaker.RecordFailure()
			slog.Debug("Redis error during cache lookup", "error", err, "shortCode", shortCode, "circuitState", d.circuitBreaker.GetState())
		} else {
			// Cache miss (redis.Nil) - this is normal, not a failure
			d.circuitBreaker.RecordSuccess()
		}
	} else {
		// Circuit is open - skip Redis
		slog.Debug("Circuit breaker open, skipping Redis cache", "shortCode", shortCode)
	}

	// Cache miss or expired - get from database
	// Multiple concurrent requests for the same key may all hit the DB,
	// but that's acceptable - the DB connection pool handles it efficiently.
	entity, err := d.underlying.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	// Cache the result if found (non-blocking - failures don't affect read)
	if entity != nil && !d.circuitBreaker.IsOpen() {
		if err := d.setCache(ctx, entity); err != nil {
			d.circuitBreaker.RecordFailure()
			slog.Debug("Failed to cache retrieved record", "error", err, "shortCode", shortCode, "circuitState", d.circuitBreaker.GetState())
		} else {
			d.circuitBreaker.RecordSuccess()
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
