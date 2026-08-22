package service

import (
	"github.com/jb843051627/avalanche-net/internal/engine"
	"github.com/jb843051627/avalanche-net/internal/model"
)

// RecordObservation 登记雪崩目击记录。
func (s *Service) RecordObservation(o *model.Observation) error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.StationID != "" {
		if _, err := s.store.GetStation(o.StationID); err != nil {
			return err
		}
	}
	o.ObservedAt = o.ObservedAt.UTC()
	s.met.Inc("observation.recorded")
	return s.store.InsertObservation(o)
}

// GetObservation 查询目击记录。
func (s *Service) GetObservation(id string) (*model.Observation, error) {
	return s.store.GetObservation(id)
}

// ListObservations 返回区域目击列表。
func (s *Service) ListObservations(regionID string) ([]*model.Observation, error) {
	return s.store.ListObservationsByRegion(regionID)
}

// ObservationStats 区域触发方式统计。
func (s *Service) ObservationStats(regionID string) (map[string]int, error) {
	return s.store.CountObservationsByTrigger(regionID)
}

// CompareStationProfiles 对比站点两次剖面的演化。
// fromID/toID 均须存在且 to 观测时间不早于 from。
func (s *Service) CompareStationProfiles(fromID, toID string) (*engine.ProfileDiff, error) {
	from, err := s.store.GetProfile(fromID)
	if err != nil {
		return nil, err
	}
	to, err := s.store.GetProfile(toID)
	if err != nil {
		return nil, err
	}
	if to.ObservedAt.Before(from.ObservedAt) {
		from, to = to, from
	}
	diff := engine.CompareProfiles(from, to)
	s.met.Inc("profile.compared")
	return &diff, nil
}

// BuildRegionRose 聚合区域各站最新评估为危险玫瑰图。
func (s *Service) BuildRegionRose(regionID string) ([]engine.RoseCell, error) {
	stations, err := s.store.ListStations(regionID)
	if err != nil {
		return nil, err
	}
	entries := make([]engine.RoseEntry, 0, len(stations))
	for _, st := range stations {
		evs, err := s.store.LatestEvaluations(st.ID, 1)
		if err != nil || len(evs) == 0 {
			continue
		}
		entries = append(entries, engine.RoseEntry{
			Aspect:     st.Aspect,
			ElevationM: st.ElevationM,
			Level:      evs[0].DangerLevel,
		})
	}
	cells := engine.BuildRose(entries)
	s.met.Inc("rose.built")
	return cells, nil
}

// BulletinDraftFromRose 依据玫瑰图为区域生成公报草稿建议。
func (s *Service) BulletinDraftFromRose(regionID string, issuedFor string) (*model.Bulletin, error) {
	cells, err := s.BuildRegionRose(regionID)
	if err != nil {
		return nil, err
	}
	b := &model.Bulletin{
		ID:            "bul-" + regionID + "-" + issuedFor,
		RegionID:      regionID,
		AboveTreeline: engine.MaxBandLevel(cells, "above"),
		NearTreeline:  engine.MaxBandLevel(cells, "near"),
		BelowTreeline: engine.MaxBandLevel(cells, "below"),
		Summary:       "auto draft from rose",
	}
	return b, nil
}
