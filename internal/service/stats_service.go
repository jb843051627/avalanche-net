package service

import (
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// DailySummaryRow 是站点单日采集统计行。
type DailySummaryRow struct {
	Day       string             `json:"day"`
	StationID string             `json:"station_id"`
	Readings  int                `json:"readings"`
	Alerts    int                `json:"alerts"`
	ByKind    map[string]int     `json:"by_kind"`
	MaxValues map[string]float64 `json:"max_values"`
}

// StationDailySummary 按天聚合站点读数与告警数（UTC 日界）。
func (s *Service) StationDailySummary(stationID string, from, to time.Time) ([]DailySummaryRow, error) {
	if _, err := s.store.GetStation(stationID); err != nil {
		return nil, err
	}
	rows, err := s.store.ListReadings(stationID, from.UTC(), to.UTC())
	if err != nil {
		return nil, err
	}
	alerts, err := s.store.ListAlertsByStation(stationID, "")
	if err != nil {
		return nil, err
	}
	index := map[string]*DailySummaryRow{}
	var order []string
	for _, r := range rows {
		day := r.RecordedAt.UTC().Format("2006-01")
		row := index[day]
		if row == nil {
			row = &DailySummaryRow{
				Day: day, StationID: stationID,
				ByKind:    map[string]int{},
				MaxValues: map[string]float64{},
			}
			index[day] = row
			order = append(order, day)
		}
		row.Readings++
		row.ByKind[string(r.SensorKind)]++
		if r.Value > row.MaxValues[string(r.SensorKind)] {
			row.MaxValues[string(r.SensorKind)] = r.Value
		}
	}
	for _, a := range alerts {
		day := a.TriggeredAt.UTC().Format("2006-01")
		if row := index[day]; row != nil {
			row.Alerts++
		}
	}
	out := make([]DailySummaryRow, 0, len(order))
	for i := len(order) - 1; i >= 0; i-- {
		out = append(out, *index[order[i]])
	}
	return out, nil
}

// RegionLoad 返回区域内各站点的负载概况（评估次数/活跃告警/最近评分）。
type RegionLoad struct {
	StationID     string              `json:"station_id"`
	Status        model.StationStatus `json:"status"`
	Evaluations7d int                 `json:"evaluations_7d"`
	ActiveAlerts  int                 `json:"active_alerts"`
	LatestScore   float64             `json:"latest_score"`
}

// RegionLoadRanking 计算区域负载排行（按最近评分降序）。
func (s *Service) RegionLoadRanking(regionID string) ([]RegionLoad, error) {
	stations, err := s.store.ListStations(regionID)
	if err != nil {
		return nil, err
	}
	since := s.clk.Now().UTC().Add(-7 * 24 * time.Hour)
	out := make([]RegionLoad, 0, len(stations))
	for _, st := range stations {
		row := RegionLoad{StationID: st.ID, Status: st.Status}
		if n, err := s.store.CountEvaluationsSince(since); err == nil {
			row.Evaluations7d = n
		}
		if alerts, err := s.store.ListAlertsByStation(st.ID, model.StateActive); err == nil {
			row.ActiveAlerts = len(alerts)
		}
		if evs, err := s.store.LatestEvaluations(st.ID, 1); err == nil && len(evs) > 0 {
			row.LatestScore = evs[0].Score
		}
		out = append(out, row)
	}
	sortRegionLoad(out)
	return out, nil
}

func sortRegionLoad(rows []RegionLoad) {
	for i := 1; i < len(rows); i++ {
		for j := i; j > 0 && rows[j].LatestScore > rows[j-1].LatestScore; j-- {
			rows[j], rows[j-1] = rows[j-1], rows[j]
		}
	}
}

// NetworkOverview 全网概览指标。
type NetworkOverview struct {
	Stations       int              `json:"stations"`
	Online         int              `json:"online"`
	ActiveCritical int              `json:"active_critical"`
	Evaluations24h int              `json:"evaluations_24h"`
	Metrics        map[string]int64 `json:"metrics"`
}

// Overview 汇总全网状态。
func (s *Service) Overview() (*NetworkOverview, error) {
	stations, err := s.store.ListStations("")
	if err != nil {
		return nil, err
	}
	ov := &NetworkOverview{Metrics: s.met.Snapshot()}
	ov.Stations = len(stations)
	for _, st := range stations {
		if st.Status != model.StatusOffline {
			ov.Online++
		}
	}
	if n, err := s.store.CountActiveCritical(); err == nil {
		ov.ActiveCritical = n
	}
	if n, err := s.store.CountEvaluationsSince(s.clk.Now().UTC().Add(-24 * time.Hour)); err == nil {
		ov.Evaluations24h = n
	}
	return ov, nil
}
