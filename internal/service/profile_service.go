package service

import (
	"strconv"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

func formatFloat(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// CreateProfile 保存一次人工观测的雪层剖面（校验层序后落库）。
func (s *Service) CreateProfile(p *model.SnowProfile) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if _, err := s.store.GetStation(p.StationID); err != nil {
		return err
	}
	p.ObservedAt = p.ObservedAt.UTC()
	var total float64
	for _, l := range p.Layers {
		total += l.ThicknessCm()
	}
	p.TotalCm = total
	if err := s.store.InsertProfile(p); err != nil {
		return err
	}
	s.met.Inc("profile.created")
	return nil
}

// GetProfile 查询剖面。
func (s *Service) GetProfile(id string) (*model.SnowProfile, error) {
	return s.store.GetProfile(id)
}

// LatestProfile 返回站点最近剖面。
func (s *Service) LatestProfile(stationID string) (*model.SnowProfile, error) {
	return s.store.LatestProfile(stationID)
}

// ProfileHistoryCount 返回站点历史剖面数。
func (s *Service) ProfileHistoryCount(stationID string) (int, error) {
	return s.store.CountProfilesByStation(stationID)
}

// NormalizeLayers 按深度排序并重排 index（观测员录入顺序可能乱）。
func NormalizeLayers(layers []model.SnowLayer) []model.SnowLayer {
	out := layers
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].DepthFromCm < out[j-1].DepthFromCm; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// ParseLayerRow 解析一行剖面文本："35-50,240,faceted,1finger,-2.5"。
func ParseLayerRow(row string) (model.SnowLayer, error) {
	parts := splitCSVLine(row)
	if len(parts) != 6 {
		return model.SnowLayer{}, model.ErrNoLayers
	}
	var l model.SnowLayer
	bounds := splitDash(parts[0])
	if len(bounds) != 2 {
		return model.SnowLayer{}, model.ErrLayerOrder
	}
	from, errFrom := strconv.ParseFloat(bounds[0], 64)
	to, errTo := strconv.ParseFloat(bounds[1], 64)
	density, errDensity := strconv.ParseFloat(parts[1], 64)
	temp, errTemp := strconv.ParseFloat(parts[5], 64)
	if errFrom != nil || errTo != nil || errDensity != nil || errTemp != nil {
		return model.SnowLayer{}, model.ErrReadingOutOfRange
	}
	l.DepthFromCm = from
	l.DepthToCm = to
	l.DensityKgM3 = density
	l.GrainShape = model.GrainShape(parts[2])
	l.Hardness = model.Hardness(parts[3])
	l.TempC = temp
	return l, nil
}

// ObservedWithin 判断剖面观测时间是否在窗口内。
func ObservedWithin(p *model.SnowProfile, d time.Duration, now time.Time) bool {
	return !p.ObservedAt.Before(now.Add(-d)) && !p.ObservedAt.After(now)
}
