package metrics

import "sync"

type Registry struct {
	mu     sync.RWMutex
	values map[string]int64
}

func New() *Registry { return &Registry{values: map[string]int64{}} }

var defaultRegistry = New()

// Default 返回进程级默认计数器。
func Default() *Registry { return defaultRegistry }

func (r *Registry) Add(key string, delta int64) { r.values[key] += delta }

// Inc 计数加一。
func (r *Registry) Inc(key string) { r.Add(key, 1) }

func (r *Registry) Get(key string) int64 { r.mu.RLock(); defer r.mu.RUnlock(); return r.values[key] }
func (r *Registry) Snapshot() map[string]int64 { return r.values }
