package cache

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	lock                = &sync.Mutex{}
	redisClientInstance *redis.Client
)

// Init initializes the Redis client singleton. It's safe to call multiple times.
func Init(ctx context.Context) error {
	if redisClientInstance != nil {
		return nil
	}

	lock.Lock()
	defer lock.Unlock()

	// Double-check after acquiring lock.
	if redisClientInstance != nil {
		return nil
	}

	slog.Info("Initializing Redis client")
	redisClientInstance = redis.NewClient(&redis.Options{
		Addr:         "localhost:6380",
		Password:     "",              // no password set
		DB:           0,               // use default DB
		DialTimeout:  5 * time.Second, // connection timeout
		PoolSize:     1000,            // connection pool size (increased for high concurrency)
		MinIdleConns: 100,             // minimum idle connections to keep warm (increased)
		MaxRetries:   2,               // retry failed commands
		PoolTimeout:  5 * time.Second, // timeout for getting connection from pool (increased)
		ReadTimeout:  3 * time.Second, // read timeout (increased - Redis is fast but needs buffer under load)
		WriteTimeout: 3 * time.Second, // write timeout (increased - Redis is fast but needs buffer under load)
	})

	// Test the connection.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := redisClientInstance.Ping(ctx).Err(); err != nil {
		redisClientInstance = nil
		return err
	}

	slog.Info("Redis client initialized successfully")
	return nil
}

// GetClient returns the Redis client instance, or nil if not initialized.
func GetClient() *redis.Client {
	return redisClientInstance
}
