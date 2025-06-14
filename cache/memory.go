package cache

import (
	"time"

	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/contrab/freelru"
	"github.com/sagernet/sing/contrab/maphash"
	"github.com/sagernet/srsc/adapter"
)

var _ adapter.Cache = (*MemoryCache)(nil)

type MemoryCache struct {
	freelru.Cache[string, *adapter.SavedBinary]
}

func NewMemory(timeout time.Duration) *MemoryCache {
	cache := common.Must1(freelru.NewSharded[string, *adapter.SavedBinary](1024, maphash.NewHasher[string]().Hash32))
	cache.SetLifetime(timeout)
	return &MemoryCache{
		Cache: cache,
	}
}

func (c *MemoryCache) Start() error {
	return nil
}

func (c *MemoryCache) Close() error {
	return nil
}

func (c *MemoryCache) LoadBinary(tag string) (*adapter.SavedBinary, error) {
	savedBinary, loaded := c.Get(tag)
	if !loaded {
		return nil, nil
	}
	return savedBinary, nil
}

func (c *MemoryCache) SaveBinary(tag string, binary *adapter.SavedBinary) error {
	c.Add(tag, binary)
	return nil
}
