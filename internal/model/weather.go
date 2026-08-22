package model

import "time"

// WeatherSample 是区域气象站样本，用于风载修正与降水负载评估。
type WeatherSample struct {
	RegionID     string    `json:"region_id"`
	RecordedAt   time.Time `json:"recorded_at"`
	WindSpeedKmh float64   `json:"wind_speed_kmh"`
	AirTempC     float64   `json:"air_temp_c"`
	Precip24hMm  float64   `json:"precip_24h_mm"`
	NewSnow24hCm float64   `json:"new_snow_24h_cm"`
}

// WindLoadingFactor 计算风载修正系数：
// 风速 20-60 km/h 区间输雪效应最强（系数最高），超过 80 后雪源耗尽回落。
func (w *WeatherSample) WindLoadingFactor() float64 {
	v := w.WindSpeedKmh
	switch {
	case v < 10:
		return 1.0
	case v < 20:
		return 1.05
	case v < 40:
		return 1.25
	case v < 60:
		return 1.35
	case v < 80:
		return 1.15
	default:
		return 0.95
	}
}

// SnowLoadWaterEq 把新增雪量折算成毫米水当量（新雪密度近似 100 kg/m³）。
func (w *WeatherSample) SnowLoadWaterEq() float64 {
	return w.NewSnow24hCm * 0.1
}

// RapidLoading 判断是否属于快速加载事件（24h 新雪水当量 ≥ 25mm）。
func (w *WeatherSample) RapidLoading() bool {
	return w.SnowLoadWaterEq() >= 25
}
