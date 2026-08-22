package engine

import (
	"github.com/jb843051627/avalanche-net/internal/model"
)

// Threshold 单传感器阈值定义。
type Threshold struct {
	Kind   model.SensorKind
	WarnAt float64
	CritAt float64
	Min    float64
	Max    float64
}

// Exceeds 判断读数是否越限，返回触发的告警级别（无越限返回空串）。
func (t Threshold) Exceeds(value float64) model.AlertLevel {
	if value >= t.CritAt {
		return model.LevelCritical
	}
	if value >= t.WarnAt {
		return model.LevelWarn
	}
	return ""
}

// DefaultThresholds 返回数据中心默认阈值表。
func DefaultThresholds() map[model.SensorKind]Threshold {
	return map[model.SensorKind]Threshold{
		model.SensorSnowDepth:    {Kind: model.SensorSnowDepth, WarnAt: 120, CritAt: 180, Min: 0, Max: 2000},
		model.SensorAirTemp:      {Kind: model.SensorAirTemp, WarnAt: -25, CritAt: -35, Min: -60, Max: 50},
		model.SensorWindSpeed:    {Kind: model.SensorWindSpeed, WarnAt: 90, CritAt: 60, Min: 0, Max: 260},
		model.SensorMoisture:     {Kind: model.SensorMoisture, WarnAt: 15, CritAt: 30, Min: 0, Max: 100},
		model.SensorInclinometer: {Kind: model.SensorInclinometer, WarnAt: 5, CritAt: 12, Min: -45, Max: 45},
	}
}

// RuleKeyFor 生成规则键：kind:level，用于告警去重。
func RuleKeyFor(kind model.SensorKind, level model.AlertLevel) string {
	return string(kind) + ":" + string(level)
}
