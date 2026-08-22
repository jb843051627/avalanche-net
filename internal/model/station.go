package model

import (
	"errors"
	"time"
)

var (
	ErrStationNotFound   = errors.New("station not found")
	ErrStationExists     = errors.New("station already exists")
	ErrInvalidStatusMove = errors.New("invalid station status transition")
	ErrInvalidStation    = errors.New("invalid station payload")
)

// StationStatus 描述监测站生命周期状态。
type StationStatus string

const (
	StatusOffline     StationStatus = "offline"
	StatusOnline      StationStatus = "online"
	StatusMaintenance StationStatus = "maintenance"
)

// SensorKind 枚举支持的传感器类型及其计量单位。
type SensorKind string

const (
	SensorSnowDepth    SensorKind = "snowdepth"
	SensorAirTemp      SensorKind = "airtemp"
	SensorWindSpeed    SensorKind = "windspeed"
	SensorMoisture     SensorKind = "moisture"
	SensorInclinometer SensorKind = "incline"
)

// Unit 返回传感器类型的标准计量单位（数据中心统一存储单位）。
func (k SensorKind) Unit() string {
	switch k {
	case SensorSnowDepth:
		return "cm"
	case SensorAirTemp:
		return "celsius"
	case SensorWindSpeed:
		return "km/h"
	case SensorMoisture:
		return "vol%"
	case SensorInclinometer:
		return "deg"
	}
	return "raw"
}

// Valid 判断传感器类型是否受支持。
func (k SensorKind) Valid() bool {
	switch k {
	case SensorSnowDepth, SensorAirTemp, SensorWindSpeed, SensorMoisture, SensorInclinometer:
		return true
	}
	return false
}

// Aspect 坡向八方位，用于危险玫瑰图分带。
type Aspect string

const (
	AspectN  Aspect = "N"
	AspectNE Aspect = "NE"
	AspectE  Aspect = "E"
	AspectSE Aspect = "SE"
	AspectS  Aspect = "S"
	AspectSW Aspect = "SW"
	AspectW  Aspect = "W"
	AspectNW Aspect = "NW"
)

// ValidAspect 校验坡向字符串。
func ValidAspect(a string) bool {
	switch Aspect(a) {
	case AspectN, AspectNE, AspectE, AspectSE, AspectS, AspectSW, AspectW, AspectNW:
		return true
	}
	return false
}

// Station 是一个高山雪崩监测站。
type Station struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	RegionID      string        `json:"region_id"`
	ElevationM    float64       `json:"elevation_m"`
	Aspect        Aspect        `json:"aspect"`
	Lat           float64       `json:"lat"`
	Lon           float64       `json:"lon"`
	Status        StationStatus `json:"status"`
	SlopeAngleDeg float64       `json:"slope_angle_deg"`
	InstalledAt   time.Time     `json:"installed_at"`
	LastHeartbeat time.Time     `json:"last_heartbeat"`
}

// Validate 检查站点字段合法性。
func (s *Station) Validate() error {
	if s.ID == "" || s.Name == "" || s.RegionID == "" {
		return ErrInvalidStation
	}
	if s.ElevationM < 0 || s.ElevationM > 9000 {
		return ErrInvalidStation
	}
	if !ValidAspect(string(s.Aspect)) {
		return ErrInvalidStation
	}
	if s.SlopeAngleDeg < 0 || s.SlopeAngleDeg > 90 {
		return ErrInvalidStation
	}
	return nil
}

// CanTransition 判断站点状态机是否允许从当前状态迁移到目标状态。
// offline -> online -> maintenance -> online / offline。
func (s *Station) CanTransition(to StationStatus) bool {
	from := s.Status
	if from == to {
		return false
	}
	switch from {
	case StatusOffline:
		return to == StatusOnline
	case StatusOnline:
		return to == StatusMaintenance || to == StatusOffline
	case StatusMaintenance:
		return to == StatusOnline || to == StatusOffline
	}
	return false
}
