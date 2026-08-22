package engine

import (
	"github.com/jb843051627/avalanche-net/internal/model"
)

// DangerLevelFromScore 把稳定性评分映射为危险等级。
// 0-29 low / 30-49 moderate / 50-69 considerable / 70-84 high / 85+ extreme。
func DangerLevelFromScore(score float64) model.DangerLevel {
	switch {
	case score < 25:
		return model.DangerLow
	case score < 50:
		return model.DangerModerate
	case score < 70:
		return model.DangerConsiderable
	case score < 85:
		return model.DangerHigh
	default:
		return model.DangerExtreme
	}
}

// AspectWindBoost 坡向风载加成：迎风侧（N/NE/NW）堆积更明显。
func AspectWindBoost(a model.Aspect, windFactor float64) float64 {
	switch a {
	case model.AspectN, model.AspectNW:
		return windFactor * 1.1
	case model.AspectNE:
		return windFactor * 1.05
	default:
		return windFactor
	}
}

// CombineDanger 合成剖面危险与快速加载事件的最终危险等级。
// 快速加载（24h 新雪水当量 ≥25mm）至少抬升到 considerable。
func CombineDanger(baseScore float64, w *model.WeatherSample, aspect model.Aspect) (float64, model.DangerLevel) {
	factor := 1.0
	if w != nil {
		factor = AspectWindBoost(aspect, w.WindLoadingFactor())
	}
	score := baseScore * factor
	level := DangerLevelFromScore(score)
	if score > 100 {
		score = 100
	}
	return score, level
}

// ElevationBand 按海拔返回海拔带名称（公报三带口径）。
func ElevationBand(elevationM float64) string {
	switch {
	case elevationM >= 2200:
		return "above"
	case elevationM >= 1600:
		return "near"
	default:
		return "below"
	}
}
