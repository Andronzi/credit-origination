package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/Andronzi/credit-origination/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type MemoryCache struct {
	sync.RWMutex
	items   map[string]cacheItem
	maxSize int
}

type cacheItem struct {
	value  interface{}
	expiry time.Time
}

func NewMemoryCache(maxSize int) *MemoryCache {
	return &MemoryCache{
		items:   make(map[string]cacheItem),
		maxSize: maxSize,
	}
}

func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()

	item, exists := c.items[key]
	if !exists || time.Now().After(item.expiry) {
		return nil, false
	}
	return item.value, true
}

func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.Lock()
	defer c.Unlock()

	if len(c.items) >= c.maxSize {
		for k := range c.items {
			delete(c.items, k)
			break
		}
	}

	c.items[key] = cacheItem{
		value:  value,
		expiry: time.Now().Add(ttl),
	}
}

var idempotencyCache = NewMemoryCache(10_000)

func IdempotencyInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "metadata is required")
	}

	keys := md.Get("Idempotency-key")
	if len(keys) == 0 {
		return handler(ctx, req)
	}
	key := keys[0]

	if cached, ok := idempotencyCache.Get(key); ok {
		logger.Logger.Debug("Returning cached response for idempotent request",
			zap.String("idempotency_key", key))
		return cached, nil
	}

	res, err := handler(ctx, req)
	if err == nil {
		idempotencyCache.Set(key, res, 24*time.Hour)
	}

	return res, err
}
