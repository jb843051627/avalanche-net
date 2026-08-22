package model

import (
	"errors"
	"time"
)

var (
	ErrObservationNotFound = errors.New("avalanche observation not found")
	ErrObservationInvalid  = errors.New("invalid avalanche observation")
)

// AvalancheType 雪崩类型分类。
type AvalancheType string

const (
	AvalancheSlab    AvalancheType = "slab"     // 板状雪崩
	AvalancheLoose   AvalancheType = "loose"    // 松雪雪崩
	AvalancheWetSlab AvalancheType = "wet_slab" // 湿板雪崩
	AvalancheCornice AvalancheType = "cornice"  // 雪檐崩塌
	AvalancheGlide   AvalancheType = "glide"    // 滑移雪崩
)

// AvalancheSize 欧洲雪崩规模分级（1-5）。
type AvalancheSize int

const (
	SizeSlushSmall AvalancheSize = 1
	SizeMedium     AvalancheSize = 2
	SizeLarge      AvalancheSize = 3
	SizeVeryLarge  AvalancheSize = 4
	SizeExtreme    AvalancheSize = 5
)

// TriggerType 触发方式。
type TriggerType string

const (
	TriggerNatural   TriggerType = "natural"
	TriggerHuman     TriggerType = "human"
	TriggerExplosive TriggerType = "explosive"
	TriggerUnknown   TriggerType = "unknown"
)

// Observation 是一次雪崩目击/遥感登记记录。
type Observation struct {
	ID         string        `json:"id"`
	RegionID   string        `json:"region_id"`
	StationID  string        `json:"station_id,omitempty"` // 就近监测站（可选）
	ObservedAt time.Time     `json:"observed_at"`
	Aspect     Aspect        `json:"aspect"`
	ElevationM float64       `json:"elevation_m"`
	Type       AvalancheType `json:"type"`
	Size       AvalancheSize `json:"size"`
	Trigger    TriggerType   `json:"trigger"`
	Reporter   string        `json:"reporter"`
	Comment    string        `json:"comment,omitempty"`
}

// Validate 校验目击登记字段。
func (o *Observation) Validate() error {
	if o.ID == "" || o.RegionID == "" || o.Reporter == "" {
		return ErrObservationInvalid
	}
	if !ValidAspect(string(o.Aspect)) {
		return ErrObservationInvalid
	}
	if o.ElevationM < 0 || o.ElevationM > 9000 {
		return ErrObservationInvalid
	}
	switch o.Type {
	case AvalancheSlab, AvalancheLoose, AvalancheWetSlab, AvalancheCornice, AvalancheGlide:
	default:
		return ErrObservationInvalid
	}
	if o.Size < SizeSlushSmall || o.Size > SizeExtreme {
		return ErrObservationInvalid
	}
	switch o.Trigger {
	case TriggerNatural, TriggerHuman, TriggerExplosive, TriggerUnknown:
	default:
		return ErrObservationInvalid
	}
	return nil
}
