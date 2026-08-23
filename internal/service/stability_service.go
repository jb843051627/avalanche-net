package service

import (
	"context"
	"time"

	"github.com/jb843051627/avalanche-net/internal/engine"
	"github.com/jb843051627/avalanche-net/internal/model"
)

// ProfileFreshness 评估允许使用的剖面最长观测龄期。
func (s *Service) ProfileFreshness() time.Duration { return 72 * time.Hour }

// RunEvaluation 对站点执行一次稳定性评估：
// 加载最近剖面与区域气象样本 -> 引擎评分 -> 风载合成 -> 持久化 -> 按需触发告警。
func (s *Service) RunEvaluation(ctx context.Context, stationID string) (*model.StabilityEvaluation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	st, err := s.store.GetStation(stationID)
	if err != nil {
		return nil, err
	}
	profile, err := s.store.LatestProfile(stationID)
	if err != nil {
		return nil, err
	}
	now := s.clk.Now().UTC()
	if !ObservedWithin(profile, s.ProfileFreshness(), now) {
		return nil, model.ErrReadingBadTime
	}
	baseScore, weak, hasWeak := engine.ProfileStability(profile)
	windFactor := 1.0
	var weather *model.WeatherSample
	if w, ok := s.cache.LatestByKind(stationID, model.SensorWindSpeed); ok {
		sample := &model.WeatherSample{
			RegionID:     st.RegionID,
			RecordedAt:   now,
			WindSpeedKmh: w.Value,
			AirTempC:     tempOr(s.cache, stationID),
			NewSnow24hCm: snowOr(s.cache, stationID),
		}
		weather = sample
	}
	score, level := engine.CombineDanger(baseScore, weather, st.Aspect)
	if weather != nil {
		windFactor = engine.AspectWindBoost(st.Aspect, weather.WindLoadingFactor())
	}
	ev := &model.StabilityEvaluation{
		StationID:        stationID,
		Score:            round2(score),
		DangerLevel:      level,
		WeakLayerIdx:     0,
		WeakLayerDepthCm: 0,
		WindFactor:       round2(windFactor),
		CreatedAt:        now,
	}
	if hasWeak {
		ev.WeakLayerIdx = weak.Index
		ev.WeakLayerDepthCm = weak.DepthCm
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := s.store.InsertEvaluation(ev); err != nil {
		return nil, err
	}
	s.met.Inc("evaluation.run")
	if level.Rank() >= model.DangerHigh.Rank() {
		_, _ = s.raiseAlert(context.Background(), stationID, engine.AlertCandidate{
			RuleKey: "stability:" + string(level),
			Level:   model.LevelCritical,
			Reason:  "profile stability score " + formatFloat(ev.Score) + " reached danger level " + string(level),
			Value:   ev.Score,
		}, now)
	}
	return ev, nil
}

func tempOr(c interface {
	LatestByKind(string, model.SensorKind) (model.Reading, bool)
}, stationID string) float64 {
	if r, ok := c.LatestByKind(stationID, model.SensorAirTemp); ok {
		return r.Value
	}
	return -5
}

func snowOr(c interface {
	LatestByKind(string, model.SensorKind) (model.Reading, bool)
}, stationID string) float64 {
	if r, ok := c.LatestByKind(stationID, model.SensorSnowDepth); ok {
		return r.Value / 10 // 近似：以缓存读数的十分之一估计 24h 新雪
	}
	return 0
}

func round2(v float64) float64 { return float64(int(v*100+0.5)) / 100 }
