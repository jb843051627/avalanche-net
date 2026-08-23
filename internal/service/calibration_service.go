package service

import (
	"fmt"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// RecordCalibration 登记一次传感器校准，超差记录返回告警候选。
func (s *Service) RecordCalibration(stationID string, kind model.SensorKind, referenceIn, reportedOut float64, by string) (*model.CalibrationRecord, error) {
	if _, err := s.store.GetStation(stationID); err != nil {
		return nil, err
	}
	if !kind.Valid() {
		return nil, model.ErrReadingUnknownKind
	}
	id := fmt.Sprintf("cal-%s-%s-%d", stationID, kind, s.clk.Now().UTC().Unix())
	c := model.NewCalibration(id, stationID, kind, referenceIn, reportedOut, by, s.clk.Now().UTC())
	if err := s.store.InsertCalibration(c); err != nil {
		return nil, err
	}
	s.met.Inc("calibration.recorded")
	return c, nil
}

// LatestCalibration 查询最近校准。
func (s *Service) LatestCalibration(stationID string, kind model.SensorKind) (*model.CalibrationRecord, error) {
	return s.store.LatestCalibration(stationID, kind)
}

// CalibrationDue 判断传感器是否需要复校（距上次校准超过 maxAge）。
func (s *Service) CalibrationDue(stationID string, kind model.SensorKind, maxAge time.Duration) (bool, error) {
	c, err := s.store.LatestCalibration(stationID, kind)
	if err == model.ErrCalibrationNotFound {
		return true, nil // 从未校准视为到期
	}
	if err != nil {
		return false, err
	}
	return s.clk.Now().UTC().Sub(c.CalibratedAt) > maxAge, nil
}

// OverdueSensors 扫描站点全部传感器的校准状态，返回到期清单。
func (s *Service) OverdueSensors(stationID string, maxAge time.Duration) ([]model.SensorKind, error) {
	cfgs, err := s.store.ListSensorsByStation(stationID)
	if err != nil {
		return nil, err
	}
	var due []model.SensorKind
	for _, cfg := range cfgs {
		ok, err := s.CalibrationDue(stationID, cfg.Kind, maxAge)
		if err != nil {
			continue
		}
		if ok {
			due = append(due, cfg.Kind)
		}
	}
	return due, nil
}

// CorrectReading 用最近合格校准修正读数（入库前调用）。
func (s *Service) CorrectReading(r *model.Reading) (float64, error) {
	return s.store.ApplyOffsetToReading(r.StationID, r)
}

// DriftReport 是站点漂移巡检报告行。
type DriftReport struct {
	StationID    string  `json:"station_id"`
	SensorKind   string  `json:"sensor_kind"`
	DriftPct     float64 `json:"drift_pct"`
	Acceptable   bool    `json:"acceptable"`
	CalibratedAt string  `json:"calibrated_at"`
}

// StationDriftReport 生成站点各传感器的漂移概况。
func (s *Service) StationDriftReport(stationID string) ([]DriftReport, error) {
	recs, err := s.store.ListCalibrationsByStation(stationID)
	if err != nil {
		return nil, err
	}
	seen := map[model.SensorKind]bool{}
	out := make([]DriftReport, 0, len(recs))
	for _, c := range recs {
		if seen[c.SensorKind] {
			continue
		}
		seen[c.SensorKind] = true
		out = append(out, DriftReport{
			StationID:    stationID,
			SensorKind:   string(c.SensorKind),
			DriftPct:     c.DriftPct,
			Acceptable:   c.Acceptable(),
			CalibratedAt: c.CalibratedAt.Format(time.RFC3339),
		})
	}
	return out, nil
}
