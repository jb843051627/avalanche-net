package service

import (
	"context"
	"log"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// StartHeartbeatWatcher 启动心跳守护 goroutine：
// 每个周期扫描一次站点，把心跳超时的在线站批量置离线。
// 返回停止函数；进程退出前必须调用以避免 goroutine 泄漏。
func (s *Service) StartHeartbeatWatcher(ctx context.Context) func() {
	interval := time.Duration(s.heartbeatIntervalMinutes) * time.Minute
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	ticker := time.NewTicker(interval)
	done := make(chan struct{})
	go s.heartbeatLoop(ctx, ticker, done)
	return func() {
		close(done)
	}
}

func (s *Service) heartbeatLoop(ctx context.Context, ticker *time.Ticker, done <-chan struct{}) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("heartbeat watcher stopped: %v", ctx.Err())
			return
		case <-done:
			return
		case <-ticker.C:
			s.sweepOnce(ctx)
		}
	}
}

// sweepOnce 执行一轮离线判定：last_heartbeat 早于阈值的在线站置为 offline。
func (s *Service) sweepOnce(ctx context.Context) {
	cutoff := s.clk.Now().UTC().Add(-time.Duration(s.offlineAfterMinutes) * time.Minute)
	n, err := s.store.MarkStationsOfflineBefore(cutoff)
	if err != nil {
		log.Printf("heartbeat sweep failed: %v", err)
		return
	}
	if n > 0 {
		s.met.Add("station.auto_offline", n)
	}
}

// SweepNow 手动触发一轮心跳扫描（运维接口/测试用）。
func (s *Service) SweepNow() int64 {
	cutoff := s.clk.Now().UTC().Add(-time.Duration(s.offlineAfterMinutes) * time.Minute)
	n, err := s.store.MarkStationsOfflineBefore(cutoff)
	if err != nil {
		return 0
	}
	if n > 0 {
		s.met.Add("station.auto_offline", n)
	}
	return n
}

// OfflineStations 返回当前处于离线状态的站点 ID 列表。
func (s *Service) OfflineStations() ([]string, error) {
	stations, err := s.store.ListStations("")
	if err != nil {
		return nil, err
	}
	var out []string
	for _, st := range stations {
		if st.Status == model.StatusOffline {
			out = append(out, st.ID)
		}
	}
	return out, nil
}
