package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/frangi01/bbtelgo/internal/config"
	"github.com/redis/go-redis/v9"
)

// Local Wrapper
type CacheClient struct {
	RDB *redis.Client
}

// Constructor: Creates the client and pings with a timeout.
func NewCacheClient(ctx context.Context, config config.RedisCfg) (*CacheClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &CacheClient{RDB: rdb}, nil
}

func (c *CacheClient) Close() error {
	return c.RDB.Close()
}

// SetString saves a string with an optional TTL (0 = no expiration).
func (c *CacheClient) SetString(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.RDB.Set(ctx, key, value, ttl).Err()
}

// GetString returns ("", nil) if the key does not exist.
func (c *CacheClient) GetString(ctx context.Context, key string) (string, error) {
	val, err := c.RDB.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil // key not found
	}
	return val, err
}

// helper for JSON-object
func (c *CacheClient) SetJSON(ctx context.Context, key string, v any, ttl time.Duration) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.RDB.Set(ctx, key, b, ttl).Err()
}

func (c *CacheClient) GetJSON(ctx context.Context, key string, out any) (bool, error) {
	b, err := c.RDB.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(b, out)
}


// Delete: delete one or more keys. Returns how many were removed.
func (c *CacheClient) Delete(ctx context.Context, keys ...string) (int64, error) {
	n, err := c.RDB.Del(ctx, keys...).Result()
	return n, err
}

// Exists: returns how many keys exist among those passed.
func (c *CacheClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.RDB.Exists(ctx, keys...).Result()
}

// Expire: sets TTL on a key. Returns true if the TTL is set.
func (c *CacheClient) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return c.RDB.Expire(ctx, key, ttl).Result()
}

// TTL: returns the remaining TTL (or -1 if it has no expiration, -2 if it does not exist).
func (c *CacheClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.RDB.TTL(ctx, key).Result()
}

// SetNX: sets only if the key does not exist (typical for lock/once). true=set.
func (c *CacheClient) SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	return c.RDB.SetNX(ctx, key, value, ttl).Result()
}

// IncrBy/DecrBy: atomic counters.
func (c *CacheClient) IncrBy(ctx context.Context, key string, n int64) (int64, error) {
	return c.RDB.IncrBy(ctx, key, n).Result()
}
func (c *CacheClient) DecrBy(ctx context.Context, key string, n int64) (int64, error) {
	return c.RDB.DecrBy(ctx, key, n).Result()
}

// MSet/MGet: set/get multiple (MGet returns a slice with nil for missing keys).
func (c *CacheClient) MSet(ctx context.Context, kv map[string]any) error {
	args := make([]any, 0, len(kv)*2)
	for k, v := range kv {
		args = append(args, k, v)
	}
	return c.RDB.MSet(ctx, args...).Err()
}
func (c *CacheClient) MGet(ctx context.Context, keys ...string) ([]any, error) {
	return c.RDB.MGet(ctx, keys...).Result()
}

// Hash helpers: useful for "field" objects.
func (c *CacheClient) HSet(ctx context.Context, key string, fields map[string]any) (int64, error) {
	return c.RDB.HSet(ctx, key, fields).Result()
}
func (c *CacheClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.RDB.HGetAll(ctx, key).Result()
}

// Pipeline helper: executes multiple commands in a round-trip.
func (c *CacheClient) WithPipeline(ctx context.Context, fn func(redis.Pipeliner) error) error {
	pipe := c.RDB.Pipeline()
	if err := fn(pipe); err != nil {
		return err
	}
	_, err := pipe.Exec(ctx)
	return err
}

// --- RATE LIMITING ---

// RateLimitFixedWindow implements a rate limit for a fixed window.
// Example: limit=5, window=1*time.Minute
// Returns: allowed, remaining, resetAt
func (c *CacheClient) RateLimitFixedWindow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Time, error) {
	now := time.Now()
	windowStart := now.Truncate(window)
	resetAt := windowStart.Add(window)

	// Key for this window
	k := fmt.Sprintf("rl:%s:%d", key, windowStart.Unix())

	pipe := c.RDB.TxPipeline()
	cnt := pipe.Incr(ctx, k)
	pipe.ExpireAt(ctx, k, resetAt) // exact expiration at the end of the window
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, resetAt, err
	}

	cur := int(cnt.Val())
	allowed := cur <= limit
	remaining := limit - cur
	if remaining < 0 {
		remaining = 0
	}
	return allowed, remaining, resetAt, nil
}

// RateLimitSlidingWindow: precise on a sliding window using ZSET.
// Returns: allowed, remaining, resetAt (estimate of reset based on recent events)
func (c *CacheClient) RateLimitSlidingWindow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Time, error) {
	now := time.Now()
	nowMs := now.UnixMilli()
	agoMs := now.Add(-window).UnixMilli()
	k := "rl:sw:" + key

	// Unique member (timestamp-ms + random) to avoid collisions
	member := fmt.Sprintf("%d-%d", nowMs, time.Now().UnixNano())

	pipe := c.RDB.TxPipeline()
	pipe.ZAdd(ctx, k, redis.Z{Score: float64(nowMs), Member: member})
	pipe.ZRemRangeByScore(ctx, k, "0", fmt.Sprintf("%d", agoMs)) // remove off window
	card := pipe.ZCard(ctx, k)
	pipe.Expire(ctx, k, window*2) // just in case
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, err
	}

	cur := int(card.Val())
	allowed := cur <= limit
	remaining := limit - cur
	if remaining < 0 {
		remaining = 0
	}

	// Reset estimate: when the oldest among the last "limit" events will occur
	resetAt := now.Add(window)
	return allowed, remaining, resetAt, nil
}

// --- DISTRIBUTED LOCK (safe with Lua) ---

var unlockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
  return redis.call("del", KEYS[1])
else
  return 0
end`)

// AcquireLock tries to acquire a lock with a value (token) and TTL
// true if obtained, false if already locked.
func (c *CacheClient) AcquireLock(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	// SET key value NX PX ttl
	ok, err := c.RDB.SetNX(ctx, key, value, ttl).Result()
	return ok, err
}

// ReleaseLock releases the lock only if the value matches (safe).
func (c *CacheClient) ReleaseLock(ctx context.Context, key, value string) (bool, error) {
	n, err := unlockScript.Run(ctx, c.RDB, []string{key}, value).Int()
	return n == 1, err
}

// --- SCAN / DELETE by prefix ---

// ScanPrefix returns keys with a certain prefix (uses SCAN, does not block).
func (c *CacheClient) ScanPrefix(ctx context.Context, prefix string, count int64) ([]string, error) {
	var (
		cursor uint64
		keys   []string
		err    error
	)
	pat := prefix + "*"
	for {
		var batch []string
		batch, cursor, err = c.RDB.Scan(ctx, cursor, pat, count).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, batch...)
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}

// DeleteByPrefix deletes all keys with a prefix (caution in production).
func (c *CacheClient) DeleteByPrefix(ctx context.Context, prefix string, count int64) (int64, error) {
	keys, err := c.ScanPrefix(ctx, prefix, count)
	if err != nil {
		return 0, err
	}
	if len(keys) == 0 {
		return 0, nil
	}
	return c.RDB.Del(ctx, keys...).Result()
}


/*

	// String + TTL
	if err := c.SetString(ctx, "hello", "world", 30*time.Second); err != nil {
		return err
	}
	if v, _ := c.GetString(ctx, "hello"); v == "" {
		return fmt.Errorf("get hello failed")
	}

	// JSON
	type U struct{ ID int; Name string }
	if err := c.SetJSON(ctx, "user:1", U{ID: 1, Name: "Alice"}, time.Minute); err != nil {
		return err
	}
	var u U
	ok, err := c.GetJSON(ctx, "user:1", &u)
	if err != nil || !ok || u.Name != "Alice" {
		return fmt.Errorf("json roundtrip failed: %v", err)
	}

	// Hash
	if _, err := c.HSet(ctx, "user:2", map[string]any{"name": "Bob", "age": 33}); err != nil {
		return err
	}
	if m, _ := c.HGetAll(ctx, "user:2"); m["name"] != "Bob" {
		return fmt.Errorf("hash read failed")
	}

	// Counter
	if _, err := c.IncrBy(ctx, "cnt", 1); err != nil {
		return err
	}
	if _, err := c.DecrBy(ctx, "cnt", 1); err != nil {
		return err
	}

	// Expire/TTL
	if ok, err := c.Expire(ctx, "hello", 10*time.Second); err != nil || !ok {
		return fmt.Errorf("expire failed: %v", err)
	}
	if ttl, _ := c.TTL(ctx, "hello"); ttl <= 0 {
		return fmt.Errorf("ttl not set")
	}

	// Pipeline
	if err := c.WithPipeline(ctx, func(p redis.Pipeliner) error {
		p.Set(ctx, "p:a", "1", 0)
		p.Incr(ctx, "p:cnt")
		p.Expire(ctx, "p:a", time.Minute)
		return nil
	}); err != nil {
		return err
	}

	// Rate limit (fixed)
	if allowed, _, _, err := c.RateLimitFixedWindow(ctx, "user:1:msg", 5, time.Minute); err != nil || !allowed {
		return fmt.Errorf("fixed window unexpected block: %v", err)
	}

	// Sliding window
	if allowed, _, _, err := c.RateLimitSlidingWindow(ctx, "user:1:msg", 5, time.Minute); err != nil || !allowed {
		return fmt.Errorf("sliding window unexpected block: %v", err)
	}

	// Lock distribuited
	lockKey := "locks:job42"
	token := fmt.Sprintf("%d", time.Now().UnixNano())
	ok2, err := c.AcquireLock(ctx, lockKey, token, 5*time.Second)
	if err != nil || !ok2 {
		return fmt.Errorf("acquire lock failed: %v", err)
	}
	released, err := c.ReleaseLock(ctx, lockKey, token)
	if err != nil || !released {
		return fmt.Errorf("release lock failed: %v", err)
	}

	// Scan / DeleteByPrefix
	if _, err := c.MSet(ctx, map[string]any{"pref:x": "1", "pref:y": "2"}); err != nil {
		return err
	}
	if n, err := c.DeleteByPrefix(ctx, "pref:", 1000); err != nil || n < 2 {
		return fmt.Errorf("delete by prefix failed: n=%d err=%v", n, err)
	}

	// Cleanup
	_, _ = c.Delete(ctx, "hello", "user:1", "user:2", "cnt", "p:a", "p:cnt")
	return nil
*/