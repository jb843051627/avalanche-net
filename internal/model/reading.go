package model

import (
	"errors"
	"time"
)

var (
	ErrReadingOutOfRange  = errors.New("reading value out of range")
	ErrReadingBadTime     = errors.New("reading timestamp outside ingest window")
	ErrReadingUnknownKind = errors.New("unknown sensor kind")
	ErrEmptyBatch         = errors.New("empty reading batch")
)

// Reading 是一条传感器遥测读数。数据中心统一以公制单位存储：
// snowdepth=cm、airtemp=celsius、windspeed=km/h、moisture=vol%、incline=deg。
type Reading struct {
	ID         int64      `json:"id"`
	StationID  string     `json:"station_id"`
	SensorKind SensorKind `json:"sensor_kind"`
	Value      float64    `json:"value"`
	RecordedAt time.Time  `json:"recorded_at"`
}

// ReadingBatch 是一次批量上报。
type ReadingBatch struct {
	StationID string    `json:"station_id"`
	Checksum  string    `json:"checksum"`
	Readings  []Reading `json:"readings"`
}

// ValidateRange 校验单个读数的量程（不含时间窗）。
func (r *Reading) ValidateRange() error {
	if !r.SensorKind.Valid() {
		return ErrReadingUnknownKind
	}
	switch r.SensorKind {
	case SensorSnowDepth:
		if r.Value < 0 || r.Value > 2000 {
			return ErrReadingOutOfRange
		}
	case SensorAirTemp:
		if r.Value < -60 || r.Value > 50 {
			return ErrReadingOutOfRange
		}
	case SensorWindSpeed:
		if r.Value < 0 || r.Value > 260 {
			return ErrReadingOutOfRange
		}
	case SensorMoisture:
		if r.Value < 0 || r.Value > 100 {
			return ErrReadingOutOfRange
		}
	case SensorInclinometer:
		if r.Value < -45 || r.Value > 45 {
			return ErrReadingOutOfRange
		}
	}
	return nil
}

// ValidateWindow 校验读数时间戳是否落在 [from, to] 采集窗口内。
func (r *Reading) ValidateWindow(from, to time.Time) error {
	if r.RecordedAt.Before(from) || r.RecordedAt.After(to) {
		return ErrReadingBadTime
	}
	return nil
}
