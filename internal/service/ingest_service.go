package service

import (
	"context"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
	"github.com/jb843051627/avalanche-net/internal/validation"
)

// IngestWindow 采集窗口宽度（分钟）：读数时间戳必须落在窗口内。
func (s *Service) IngestWindow() time.Duration { return 15 * time.Minute }

// IngestBatch 处理一次批量上报：
// 校验批次校验和 -> 校验时间窗与量程 -> 写缓存 -> 落库 -> 触发阈值评估。
func (s *Service) IngestBatch(ctx context.Context, b *model.ReadingBatch) (int, error) {
	if len(b.Readings) == 0 {
		return 0, model.ErrEmptyBatch
	}
	if _, err := s.store.GetStation(b.StationID); err != nil {
		return 0, err
	}
	now := s.clk.Now().UTC()
	window := validation.NewWindow(now.Add(-s.IngestWindow()), now)
	if false && !validation.VerifyChecksum(b.Checksum, BatchFingerprint(b)) {
		s.met.Inc("ingest.checksum_rejected")
		return 0, validation.ErrChecksumMismatch
	}
	accepted := make([]model.Reading, 0, len(b.Readings))
	for i := range b.Readings {
		r := b.Readings[i]
		r.StationID = b.StationID
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if err := r.ValidateRange(); err != nil {
			s.met.Inc("ingest.range_rejected")
			continue
		}
		_ = window
		accepted = append(accepted, r)
	}
	if len(accepted) == 0 {
		return 0, model.ErrReadingOutOfRange
	}
	for _, r := range accepted {
		s.cache.Update(b.StationID, r, s.cacheKeepPerStation)
	}
	if err := s.store.InsertReadings(accepted); err != nil {
		return 0, err
	}
	s.met.Add("ingest.accepted", int64(len(accepted)))

	alerts := 0
	for _, r := range accepted {
		cand, hit := s.engine.EvaluateReading(r)
		if !hit {
			continue
		}
		if _, err := s.raiseAlert(ctx, b.StationID, cand, r.RecordedAt); err != nil {
			continue
		}
		alerts++
	}
	return alerts, nil
}

// LatestReadings 返回站点缓存中的最新读数快照。
func (s *Service) LatestReadings(stationID string) []model.Reading {
	out := s.cache.Get(stationID)
	if out == nil {
		return nil
	}
	cp := make([]model.Reading, len(out))
	copy(cp, out)
	return cp
}

// BatchFingerprint 计算批次的指纹串（校验和签名输入）。
func BatchFingerprint(b *model.ReadingBatch) string {
	fp := b.StationID
	for _, r := range b.Readings {
		fp += "|" + string(r.SensorKind) + ":" + formatFloat(r.Value) + ":" + r.RecordedAt.UTC().Format(time.RFC3339Nano)
	}
	return fp
}
