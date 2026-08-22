package service

import (
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// ExpectedIntervalMinutes 每站遥测期望上报周期（分钟）。
func (s *Service) ExpectedIntervalMinutes() int { return 10 }

// GapRow 是一段遥测断档。
type GapRow struct {
	StationID  string    `json:"station_id"`
	SensorKind string    `json:"sensor_kind"`
	From       time.Time `json:"from_utc"`
	To         time.Time `json:"to_utc"`
	DurationM  float64   `json:"duration_minutes"`
}

// DetectTelemetryGaps 扫描站点窗口内各传感器的上报断档：
// 相邻读数间隔超过 2.5 倍期望周期即记为一次断档；窗口头部/尾部缺失也计入。
func (s *Service) DetectTelemetryGaps(stationID string, from, to time.Time) ([]GapRow, error) {
	if _, err := s.store.GetStation(stationID); err != nil {
		return nil, err
	}
	readings, err := s.store.ListReadings(stationID, from.UTC(), to.UTC())
	if err != nil {
		return nil, err
	}
	expected := time.Duration(float64(s.ExpectedIntervalMinutes())*2.5*60) * time.Second
	gapLimit := expected
	byKind := map[model.SensorKind][]model.Reading{}
	for _, r := range readings {
		byKind[r.SensorKind] = append(byKind[r.SensorKind], r)
	}
	var out []GapRow
	for kind, rs := range byKind {
		if len(rs) == 0 {
			continue
		}
		prevAt := from.UTC()
		for _, r := range rs {
			d := r.RecordedAt.Sub(prevAt)
			if d > gapLimit {
				out = append(out, newGap(stationID, kind, prevAt, r.RecordedAt))
			}
			prevAt = r.RecordedAt
		}
		if d := to.UTC().Sub(prevAt); d > gapLimit {
			out = append(out, newGap(stationID, kind, prevAt, to.UTC()))
		}
	}
	sortGaps(out)
	return out, nil
}

func newGap(station string, kind model.SensorKind, from, to time.Time) GapRow {
	return GapRow{
		StationID:  station,
		SensorKind: string(kind),
		From:       from,
		To:         to,
		DurationM:  to.Sub(from).Minutes(),
	}
}

func sortGaps(gaps []GapRow) {
	for i := 1; i < len(gaps); i++ {
		for j := i; j > 0 && gaps[j].DurationM > gaps[j-1].DurationM; j-- {
			gaps[j], gaps[j-1] = gaps[j-1], gaps[j]
		}
	}
}

// StationCompleteness 计算站点采集完整率（0-100）：
// 完整率 = 1 - 断档总时长 / 窗口时长，多传感器取均值。
func (s *Service) StationCompleteness(stationID string, from, to time.Time) (float64, error) {
	gaps, err := s.DetectTelemetryGaps(stationID, from, to)
	if err != nil {
		return 0, err
	}
	window := to.Sub(from).Minutes()
	if window <= 0 {
		return 0, model.ErrReadingBadTime
	}
	kinds := map[string]bool{}
	for _, g := range gaps {
		kinds[g.SensorKind] = true
	}
	station, err := s.store.GetStation(stationID)
	if err != nil {
		return 0, err
	}
	sensorCount := len(kinds)
	if sensorCount == 0 {
		cfgs, _ := s.store.ListSensorsByStation(stationID)
		if len(cfgs) > 0 {
			sensorCount = len(cfgs)
		} else {
			sensorCount = 1
		}
	}
	lostPerSensor := make([]float64, 0, sensorCount)
	_ = station
	for kind := range kinds {
		lost := 0.0
		for _, g := range gaps {
			if g.SensorKind == kind {
				lost += g.DurationM
			}
		}
		lostPerSensor = append(lostPerSensor, lost)
	}
	total := 0.0
	for _, lost := range lostPerSensor {
		ratio := lost / window
		if ratio > 1 {
			ratio = 1
		}
		total += (1 - ratio) * 100
	}
	return round2(total / float64(sensorCount)), nil
}
