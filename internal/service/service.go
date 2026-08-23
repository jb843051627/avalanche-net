package service

import (
	"github.com/jb843051627/avalanche-net/internal/cache"
	"github.com/jb843051627/avalanche-net/internal/clock"
	"github.com/jb843051627/avalanche-net/internal/engine"
	"github.com/jb843051627/avalanche-net/internal/ingest"
	"github.com/jb843051627/avalanche-net/internal/metrics"
	"github.com/jb843051627/avalanche-net/internal/store"
)

// Service 聚合全部业务依赖，是 HTTP 层的唯一入口。
type Service struct {
	store  *store.Store
	cache  *cache.ReadingCache
	engine *engine.RuleEngine
	queue  *ingest.Queue
	met    *metrics.Registry
	clk    clock.Clock

	heartbeatIntervalMinutes int
	offlineAfterMinutes      int
	cacheKeepPerStation      int
}

// Option 允许装配时微调服务参数。
type Option func(*Service)

// WithHeartbeatConfig 设置心跳巡检周期与离线判定阈值（分钟）。
func WithHeartbeatConfig(interval, offlineAfter int) Option {
	return func(s *Service) {
		if interval > 0 {
			s.heartbeatIntervalMinutes = interval
		}
		if offlineAfter > 0 {
			s.offlineAfterMinutes = offlineAfter
		}
	}
}

// WithCacheKeep 设置每站缓存保留条数。
func WithCacheKeep(n int) Option {
	return func(s *Service) {
		if n > 0 {
			s.cacheKeepPerStation = n
		}
	}
}

// New 装配服务。
func New(st *store.Store, opts ...Option) *Service {
	s := &Service{
		store:                    st,
		cache:                    cache.New(),
		engine:                   engine.NewRuleEngine(),
		queue:                    ingest.New(64),
		met:                      metrics.Default(),
		clk:                      clock.System{},
		heartbeatIntervalMinutes: 5,
		offlineAfterMinutes:      30,
		cacheKeepPerStation:      512,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Close 释放后台资源。
func (s *Service) Close() {
	s.queue.Close()
}

// Store 暴露仓储层（导出/统计等只读场景使用）。
func (s *Service) Store() *store.Store { return s.store }

// Cache 暴露读数缓存。
func (s *Service) Cache() *cache.ReadingCache { return s.cache }
