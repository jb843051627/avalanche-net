package model

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrCalibrationNotFound = errors.New("calibration record not found")
	ErrDriftOutOfRange     = errors.New("calibration drift beyond acceptance")
)

// CalibrationRecord 是一次传感器校准记录：
// 校准员在现场施加已知输入，记录仪表读数偏差。
type CalibrationRecord struct {
	ID           string     `json:"id"`
	StationID    string     `json:"station_id"`
	SensorKind   SensorKind `json:"sensor_kind"`
	ReferenceIn  float64    `json:"reference_input"` // 标准输入值
	ReportedOut  float64    `json:"reported_output"` // 仪表输出值
	Offset       float64    `json:"offset"`          // 输出-输入，读数修正量
	DriftPct     float64    `json:"drift_pct"`       // 相对满量程的漂移百分比
	CalibratedBy string     `json:"calibrated_by"`
	CalibratedAt time.Time  `json:"calibrated_at"`
}

// FullScale 返回传感器类型的满量程宽度，用于漂移归一化。
func FullScale(k SensorKind) float64 {
	switch k {
	case SensorSnowDepth:
		return 2000
	case SensorAirTemp:
		return 110 // -60..50
	case SensorWindSpeed:
		return 260
	case SensorMoisture:
		return 100
	case SensorInclinometer:
		return 90 // -45..45
	}
	return 100
}

// NewCalibration 由参考输入与实际输出推导偏移与漂移。
func NewCalibration(id, stationID string, kind SensorKind, referenceIn, reportedOut float64, by string, at time.Time) *CalibrationRecord {
	offset := reportedOut - referenceIn
	drift := offset / FullScale(kind) * 100
	if drift < 0 {
		drift = -drift
	}
	return &CalibrationRecord{
		ID:           id,
		StationID:    stationID,
		SensorKind:   kind,
		ReferenceIn:  referenceIn,
		ReportedOut:  reportedOut,
		Offset:       offset,
		DriftPct:     drift,
		CalibratedBy: by,
		CalibratedAt: at.UTC(),
	}
}

// Acceptable 判断漂移是否在验收范围内（默认满量程 2%）。
func (c *CalibrationRecord) Acceptable() bool { return c.DriftPct <= 2.0 }

// String 摘要文本（日志用）。
func (c *CalibrationRecord) String() string {
	return fmt.Sprintf("calib %s/%s off=%.3f drift=%.2f%%", c.StationID, c.SensorKind, c.Offset, c.DriftPct)
}
