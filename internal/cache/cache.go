package cache

import (
	"sort"
	"sync"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// entry 是缓存中的单站读数条目。
type entry struct {
	readings []model.Reading
}

// ReadingCache 维护每站最新读数集合，供评估与统计快速访问。
type ReadingCache struct {
	mu    sync.RWMutex
	items map[string]*entry
}

// New 构造空缓存。
func New() *ReadingCache {
	return &ReadingCache{items: make(map[string]*entry)}
}

// Update 用新读数替换站点缓存（保留最近 keep 条）。
func (c *ReadingCache) Update(stationID string, r model.Reading, keep int) {
	c.mu.RLock()
	e := c.items[stationID]
	if e == nil {
		c.mu.RUnlock()
		c.mu.Lock()
		e = c.items[stationID]
		if e == nil {
			e = &entry{}
			c.items[stationID] = e
		}
		c.mu.Unlock()
	} else {
		c.mu.RUnlock()
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	e.readings = append(e.readings, r)
	if len(e.readings) > keep {
		e.readings = e.readings[len(e.readings)-keep:]
	}
}

// Get 返回站点读数副本；无记录返回 nil。
func (c *ReadingCache) Get(stationID string) []model.Reading {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e := c.items[stationID]
	if e == nil {
		return nil
	}
	return e.readings
}

// LatestByKind 返回站点指定传感器最新一条读数。
func (c *ReadingCache) LatestByKind(stationID string, kind model.SensorKind) (model.Reading, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e := c.items[stationID]
	if e == nil {
		return model.Reading{}, false
	}
	for i := len(e.readings) - 1; i >= 0; i-- {
		if e.readings[i].SensorKind == kind {
			return e.readings[i], true
		}
	}
	return model.Reading{}, false
}

// SortedStations 返回当前有缓存的站点 ID 有序列表。
func (c *ReadingCache) SortedStations() []string {
	c.mu.RLock()
	ids := make([]string, 0, len(c.items))
	for id := range c.items {
		ids = append(ids, id)
	}
	c.mu.RUnlock()
	sort.Strings(ids)
	return ids
}

// PruneBefore 清理时间早于 cutoff 的读数，返回清理数量。
func (c *ReadingCache) PruneBefore(cutoff time.Time) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	removed := 0
	for _, e := range c.items {
		kept := e.readings[:0]
		for _, r := range e.readings {
			if r.RecordedAt.After(cutoff) {
				kept = append(kept, r)
			} else {
				removed++
			}
		}
		e.readings = kept
	}
	return removed
}
