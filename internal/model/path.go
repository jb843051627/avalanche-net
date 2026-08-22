package model

import (
	"errors"
	"time"
)

var (
	ErrPathNotFound = errors.New("avalanche path not found")
	ErrPathInvalid  = errors.New("invalid avalanche path")
)

// AvalanchePath 是一条登记在册的雪崩路径（释放区-运动区-堆积区）。
type AvalanchePath struct {
	ID            string     `json:"id"`
	RegionID      string     `json:"region_id"`
	Name          string     `json:"name"`
	StartElevM    float64    `json:"start_elevation_m"`
	EndElevM      float64    `json:"end_elevation_m"`
	Aspect        Aspect     `json:"aspect"`
	SlopeDeg      float64    `json:"slope_deg"`
	LengthM       float64    `json:"length_m"`
	HitsRoad      bool       `json:"hits_road"`       // 是否威胁道路
	HitsStructs   bool       `json:"hits_structures"` // 是否威胁建筑
	LastEventAt   *time.Time `json:"last_event_at,omitempty"`
	EventCount12m int        `json:"event_count_12m"`
	RegisteredAt  time.Time  `json:"registered_at"`
}

// Validate 校验路径字段。
func (p *AvalanchePath) Validate() error {
	if p.ID == "" || p.RegionID == "" || p.Name == "" {
		return ErrPathInvalid
	}
	if !ValidAspect(string(p.Aspect)) {
		return ErrPathInvalid
	}
	if p.StartElevM <= p.EndElevM || p.EndElevM < 0 {
		return ErrPathInvalid // 释放区必须高于堆积区
	}
	if p.SlopeDeg < 25 || p.SlopeDeg > 60 {
		return ErrPathInvalid // 释放区坡度经验区间
	}
	if p.LengthM <= 0 {
		return ErrPathInvalid
	}
	return nil
}

// VerticalDrop 返回路径落差。
func (p *AvalanchePath) VerticalDrop() float64 { return p.StartElevM - p.EndElevM }

// ExposureBase 基础暴露分：威胁道路 +40，威胁建筑 +30，两者叠加。
func (p *AvalanchePath) ExposureBase() int {
	score := 10
	if p.HitsRoad {
		score += 40
	}
	if p.HitsStructs {
		score += 30
	}
	return score
}

// ActivityBonus 活跃度加成：近一年事件数 ×5，封顶 20。
func (p *AvalanchePath) ActivityBonus() int {
	b := p.EventCount12m * 5
	if b > 20 {
		b = 20
	}
	return b
}
