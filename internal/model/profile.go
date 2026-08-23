package model

import (
	"errors"
	"time"
)

var (
	ErrProfileNotFound = errors.New("snow profile not found")
	ErrNoLayers        = errors.New("profile has no layers")
	ErrLayerOrder      = errors.New("layer depth ranges must ascend without gaps")
)

// GrainShape 描述雪晶体形态，弱层判定时按晶型加权。
type GrainShape string

const (
	GrainRounded     GrainShape = "rounded"
	GrainFaceted     GrainShape = "faceted" // 长期深霜，典型弱层
	GrainDepthHoar   GrainShape = "depth_hoar"
	GrainDecomposing GrainShape = "decomposing"
	GrainMeltFreeze  GrainShape = "melt_freeze"
)

// WeakLayerWeight 返回晶型的弱层权重（越大越危险）。
func (g GrainShape) WeakLayerWeight() float64 {
	switch g {
	case GrainFaceted:
		return 0.8
	case GrainDepthHoar:
		return 1.0
	case GrainMeltFreeze:
		return 0.5
	case GrainDecomposing:
		return 0.4
	case GrainRounded:
		return 0.1
	}
	return 0.3
}

// Hardness 手持硬度计刻度（fist 到 ice）。
type Hardness string

const (
	HardnessFist  Hardness = "fist"
	Hardness4F    Hardness = "4finger"
	Hardness1F    Hardness = "1finger"
	HardnessPenc  Hardness = "pencil"
	HardnessKnife Hardness = "knife"
	HardnessIce   Hardness = "ice"
)

// Score 返回硬度分值（1 最软、6 最硬）。
func (h Hardness) Score() int {
	switch h {
	case HardnessFist:
		return 1
	case Hardness4F:
		return 2
	case Hardness1F:
		return 3
	case HardnessPenc:
		return 4
	case HardnessKnife:
		return 5
	case HardnessIce:
		return 6
	}
	return 3
}

// SnowLayer 是积雪剖面中的一层。depth_from/depth_to 以厘米表示，
// 从雪面向下递增（depth_from < depth_to）。
type SnowLayer struct {
	Index       int        `json:"index"`
	DepthFromCm float64    `json:"depth_from_cm"`
	DepthToCm   float64    `json:"depth_to_cm"`
	DensityKgM3 float64    `json:"density_kg_m3"`
	GrainShape  GrainShape `json:"grain_shape"`
	Hardness    Hardness   `json:"hardness"`
	TempC       float64    `json:"temp_c"`
}

// ThicknessCm 返回该层厚度。
func (l SnowLayer) ThicknessCm() float64 { return l.DepthToCm - l.DepthFromCm }

// TempGradient 计算层内温度梯度（°C/m），剖面稳定性判据之一。
// 相邻层温差近似摊到本层厚度上。
func (l SnowLayer) TempGradient(prevTemp float64) float64 {
	thick := l.ThicknessCm()
	if thick <= 0 {
		return 0
	}
	dT := l.TempC - prevTemp
	if dT < 0 {
		dT = -dT
	}
	return dT / (thick / 100.0)
}

// SnowProfile 是一次人工观测的积雪层剖面记录。
type SnowProfile struct {
	ID         string      `json:"id"`
	StationID  string      `json:"station_id"`
	ObservedAt time.Time   `json:"observed_at"`
	Observer   string      `json:"observer"`
	TotalCm    float64     `json:"total_depth_cm"`
	Layers     []SnowLayer `json:"layers"`
}

// Validate 校验剖面：至少一层且深度区间单调不重叠。
func (p *SnowProfile) Validate() error {
	if p.ID == "" || p.StationID == "" {
		return ErrInvalidStation
	}
	if len(p.Layers) == 0 {
		return ErrNoLayers
	}
	prevTop := 0.0
	for i, l := range p.Layers {
		if l.Index != i+1 {
			return ErrLayerOrder
		}
		if l.DepthFromCm != prevTop || l.DepthToCm <= l.DepthFromCm {
			return ErrLayerOrder
		}
		if l.DensityKgM3 < 20 || l.DensityKgM3 > 700 {
			return ErrReadingOutOfRange
		}
		prevTop = l.DepthToCm
	}
	return nil
}
