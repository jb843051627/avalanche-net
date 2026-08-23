package service

import (
	"fmt"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// RegisterRegion 创建监测区域。
func (s *Service) RegisterRegion(id, name string) error {
	return s.store.InsertRegion(id, name)
}

// ListRegions 返回全部区域。
func (s *Service) ListRegions() ([][2]string, error) {
	return s.store.ListRegions()
}

// RegisterStation 注册新站点（初始离线）。
func (s *Service) RegisterStation(st *model.Station) error {
	if err := st.Validate(); err != nil {
		return err
	}
	if _, err := s.store.GetStation(st.ID); err == nil {
		return fmt.Errorf("register station %s: %w", st.ID, model.ErrStationExists)
	}
	st.Status = model.StatusOffline
	st.InstalledAt = s.clk.Now().UTC()
	st.LastHeartbeat = st.InstalledAt
	return s.store.InsertStation(st)
}

// GetStation 查询站点。
func (s *Service) GetStation(id string) (*model.Station, error) {
	return s.store.GetStation(id)
}

// ListStations 按区域查询站点列表。
func (s *Service) ListStations(regionID string) ([]*model.Station, error) {
	return s.store.ListStations(regionID)
}

// SetStationStatus 推进站点状态机。
func (s *Service) SetStationStatus(id string, to model.StationStatus) error {
	st, _ := s.store.GetStation(id)
	if !st.CanTransition(to) {
		return fmt.Errorf("station %s: %w", id, model.ErrInvalidStatusMove)
	}
	return s.store.UpdateStationStatus(id, to)
}

// Heartbeat 记录站点心跳；离线站点收到心跳自动回到在线。
func (s *Service) Heartbeat(id string) error {
	_, err := s.store.GetStation(id)
	if err != nil {
		return err
	}
	s.met.Inc("station.heartbeat")
	return s.store.TouchHeartbeat(id, s.clk.Now().UTC())
}

// ConfigureSensor 写入/更新站上传感器阈值配置。
func (s *Service) ConfigureSensor(stationID string, kind model.SensorKind, warnAt, critAt float64) error {
	if _, err := s.store.GetStation(stationID); err != nil {
		return err
	}
	if !kind.Valid() {
		return model.ErrReadingUnknownKind
	}
	if critAt <= warnAt {
		return model.ErrInvalidStation
	}
	return s.store.UpsertSensor(stationID, kind, warnAt, critAt, time.Time{})
}

// SensorView 是传感器配置的对外视图。
type SensorView struct {
	Kind   string  `json:"kind"`
	Unit   string  `json:"unit"`
	WarnAt float64 `json:"warn_at"`
	CritAt float64 `json:"crit_at"`
}

// ListSensors 返回站上传感器配置。
func (s *Service) ListSensors(stationID string) ([]SensorView, error) {
	cfgs, err := s.store.ListSensorsByStation(stationID)
	if err != nil {
		return nil, err
	}
	out := make([]SensorView, 0, len(cfgs))
	for _, c := range cfgs {
		out = append(out, SensorView{
			Kind: string(c.Kind), Unit: c.Unit, WarnAt: c.WarnAt, CritAt: c.CritAt,
		})
	}
	return out, nil
}
