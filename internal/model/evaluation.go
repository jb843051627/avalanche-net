package model

import (
	"errors"
	"fmt"
	"time"
)

var ErrInvalidDanger = errors.New("invalid danger level")

// DangerLevel 雪崩危险等级（欧洲五级制）。
type DangerLevel string

const (
	DangerLow          DangerLevel = "low"
	DangerModerate     DangerLevel = "moderate"
	DangerConsiderable DangerLevel = "considerable"
	DangerHigh         DangerLevel = "high"
	DangerExtreme      DangerLevel = "extreme"
)

// Rank 返回危险等级序数（1-5）。
func (d DangerLevel) Rank() int {
	switch d {
	case DangerLow:
		return 1
	case DangerModerate:
		return 2
	case DangerConsiderable:
		return 3
	case DangerHigh:
		return 4
	case DangerExtreme:
		return 5
	}
	return 0
}

// ParseDangerLevel 从字符串解析危险等级。
func ParseDangerLevel(s string) (DangerLevel, error) {
	switch DangerLevel(s) {
	case DangerLow, DangerModerate, DangerConsiderable, DangerHigh, DangerExtreme:
		return DangerLevel(s), nil
	}
	return "", fmt.Errorf("%w: %q", ErrInvalidDanger, s)
}

// StabilityEvaluation 是一次剖面稳定性评估的持久化结果。
type StabilityEvaluation struct {
	ID               int64       `json:"id"`
	StationID        string      `json:"station_id"`
	Score            float64     `json:"score"` // 0(极稳) - 100(极不稳)
	DangerLevel      DangerLevel `json:"danger_level"`
	WeakLayerIdx     int         `json:"weak_layer_idx"` // 1-based，0 表示无弱层
	WeakLayerDepthCm float64     `json:"weak_layer_depth_cm"`
	WindFactor       float64     `json:"wind_factor"`
	CreatedAt        time.Time   `json:"created_at"`
}
