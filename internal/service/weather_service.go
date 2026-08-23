package service

import (
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// RecordWeatherSample 落库区域气象样本。
func (s *Service) RecordWeatherSample(w *model.WeatherSample) error {
	if w.RegionID == "" {
		return model.ErrInvalidStation
	}
	w.RecordedAt = w.RecordedAt.UTC()
	s.met.Inc("weather.recorded")
	return s.store.InsertWeatherSample(w)
}

// LatestWeather 返回区域最新气象样本（无记录返回 nil）。
func (s *Service) LatestWeather(regionID string) (*model.WeatherSample, error) {
	return s.store.LatestWeatherSample(regionID)
}

// WeatherSince 返回窗口内全部样本。
func (s *Service) WeatherSince(regionID string, since time.Time) ([]model.WeatherSample, error) {
	return s.store.ListWeatherSince(regionID, since.UTC())
}

// LoadingSummary 汇总区域 24 小时加载事件。
type LoadingSummary struct {
	RegionID     string  `json:"region_id"`
	Samples      int     `json:"samples"`
	MaxWindKmh   float64 `json:"max_wind_kmh"`
	Precip24hMm  float64 `json:"precip_24h_mm"`
	NewSnow24hCm float64 `json:"new_snow_24h_cm"`
	RapidLoading bool    `json:"rapid_loading"`
}

// RegionLoadingSummary 统计区域最近 window 时长的加载概况：
// 风速取峰值，降水与新雪取最新样本的累计值。
func (s *Service) RegionLoadingSummary(regionID string, window time.Duration) (*LoadingSummary, error) {
	since := s.clk.Now().UTC().Add(-window)
	samples, err := s.store.ListWeatherSince(regionID, since)
	if err != nil {
		return nil, err
	}
	sum := &LoadingSummary{RegionID: regionID, Samples: len(samples)}
	for _, w := range samples {
		if w.WindSpeedKmh > sum.MaxWindKmh {
			sum.MaxWindKmh = w.WindSpeedKmh
		}
	}
	last := samples[len(samples)-1]
	sum.Precip24hMm = last.Precip24hMm
	sum.NewSnow24hCm = last.NewSnow24hCm
	sum.RapidLoading = last.RapidLoading()
	return sum, nil
}

// WindFactorForRegion 计算区域当前风载修正系数（无样本时为 1.0）。
func (s *Service) WindFactorForRegion(regionID string) float64 {
	w, err := s.store.LatestWeatherSample(regionID)
	if err != nil {
		return 1.0
	}
	return w.WindLoadingFactor()
}
