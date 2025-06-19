package cache

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/sagernet/sing-box/common/tls"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/srsc/adapter"
	"github.com/sagernet/srsc/option"

	"github.com/redis/go-redis/v9"
)

var _ adapter.Cache = (*RedisCache)(nil)

type RedisCache struct {
	ctx        context.Context
	options    *redis.UniversalOptions
	tlsConfig  tls.Config
	client     redis.UniversalClient
	expiration time.Duration
}

func NewRedis(ctx context.Context, expiration time.Duration, options option.RedisCacheOptions) (*RedisCache, error) {
	var (
		address []string
		server  string
	)
	if len(options.Address) > 0 {
		address = options.Address
		if firstHost, _, err := net.SplitHostPort(options.Address[0]); err == nil {
			server = firstHost
		}
	} else {
		address = []string{"localhost:6379"}
	}
	var protocol int
	if options.Protocol != 0 {
		protocol = options.Protocol
	} else {
		protocol = 3
	}
	var stdConfig *tls.STDConfig
	if options.TLS != nil && options.TLS.Enabled {
		tlsConfig, err := tls.NewClient(ctx, server, common.PtrValueOrDefault(options.TLS))
		if err != nil {
			return nil, err
		}
		stdConfig, err = tlsConfig.Config()
		if err != nil {
			return nil, err
		}
	}
	return &RedisCache{
		ctx: ctx,
		client: redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:     address,
			Password:  options.Password,
			DB:        options.DB,
			Protocol:  protocol,
			TLSConfig: stdConfig,
			PoolSize:  options.PoolSize,
		}),
		expiration: expiration,
	}, nil
}

func (r *RedisCache) Start() error {
	return nil
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}

func (r *RedisCache) LoadBinary(tag string) (*adapter.SavedBinary, error) {
	binaryBytes, err := r.client.Get(r.ctx, tag).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	binary := &adapter.SavedBinary{}
	err = binary.UnmarshalBinary(binaryBytes)
	if err != nil {
		return nil, err
	}
	return binary, nil
}

func (r *RedisCache) SaveBinary(tag string, binary *adapter.SavedBinary) error {
	binaryBytes, err := binary.MarshalBinary()
	if err != nil {
		return err
	}
	err = r.client.Set(r.ctx, tag, binaryBytes, r.expiration).Err()
	if err != nil {
		return err
	}
	return nil
}
