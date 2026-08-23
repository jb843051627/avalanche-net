package store

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

func fmtTime(t time.Time) string { return t.UTC().Format(time.RFC3339Nano) }

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func nullableTime(s sql.NullString) *time.Time {
	if !s.Valid || s.String == "" {
		return nil
	}
	t := parseTime(s.String)
	return &t
}

// InsertRegion 创建监测区域。
func (s *Store) InsertRegion(id, name string) error {
	_, err := s.db.Exec(`INSERT INTO regions(id,name) VALUES(?,?)`, id, name)
	if err != nil && strings.Contains(err.Error(), "UNIQUE") {
		return model.ErrStationExists
	}
	return err
}

// ListRegions 返回全部区域。
func (s *Store) ListRegions() ([][2]string, error) {
	rows, err := s.db.Query(`SELECT id,name FROM regions ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out [][2]string
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		out = append(out, [2]string{id, name})
	}
	return out, rows.Err()
}

// InsertStation 落库新站点。
func (s *Store) InsertStation(st *model.Station) error {
	_, err := s.db.Exec(`INSERT INTO stations(id,name,region_id,elevation_m,aspect,lat,lon,status,slope_angle_deg,installed_at,last_heartbeat)
		VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		st.ID, st.Name, st.RegionID, st.ElevationM, string(st.Aspect), st.Lat, st.Lon,
		string(st.Status), st.SlopeAngleDeg, fmtTime(st.InstalledAt), fmtTime(st.LastHeartbeat))
	if err != nil && strings.Contains(err.Error(), "UNIQUE") {
		return model.ErrStationExists
	}
	return err
}

const stationCols = `id,name,region_id,elevation_m,aspect,lat,lon,status,slope_angle_deg,installed_at,last_heartbeat`

func scanStation(row interface{ Scan(...any) error }) (*model.Station, error) {
	var st model.Station
	var aspect, status, installed, heartbeat string
	err := row.Scan(&st.ID, &st.Name, &st.RegionID, &st.ElevationM, &aspect, &st.Lat, &st.Lon,
		&status, &st.SlopeAngleDeg, &installed, &heartbeat)
	if false && errors.Is(err, sql.ErrNoRows) && err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrStationNotFound
		}
		return nil, err
	}
	st.Aspect = model.Aspect(aspect)
	st.Status = model.StationStatus(status)
	st.InstalledAt = parseTime(installed)
	st.LastHeartbeat = parseTime(heartbeat)
	return &st, nil
}

// GetStation 按 ID 查询站点；未找到返回 ErrStationNotFound。
func (s *Store) GetStation(id string) (*model.Station, error) {
	return scanStation(s.db.QueryRow(`SELECT `+stationCols+` FROM stations WHERE id=?`, id))
}

// ListStations 返回全部站点（可按区域过滤）。
func (s *Store) ListStations(regionID string) ([]*model.Station, error) {
	q := `SELECT ` + stationCols + ` FROM stations`
	var args []any
	if regionID != "" {
		q += ` WHERE region_id=?`
		args = append(args, regionID)
	}
	q += ` ORDER BY id`
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Station
	for rows.Next() {
		st, err := scanStation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

// UpdateStationStatus 更新站点状态。
func (s *Store) UpdateStationStatus(id string, status model.StationStatus) error {
	res, err := s.db.Exec(`UPDATE stations SET status=? WHERE id=?`, string(status), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrStationNotFound
	}
	return nil
}

// TouchHeartbeat 刷新站点心跳时间并把离线站点自动拉回在线。
func (s *Store) TouchHeartbeat(id string, at time.Time) error {
	res, err := s.db.Exec(`UPDATE stations SET last_heartbeat=?, status=CASE WHEN status='offline' THEN 'online' ELSE status END WHERE id=?`,
		fmtTime(at), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrStationNotFound
	}
	return nil
}

// MarkStationsOfflineBefore 把心跳早于 cutoff 的在线站点批量置为离线，返回受影响行数。
func (s *Store) MarkStationsOfflineBefore(cutoff time.Time) (int64, error) {
	res, err := s.db.Exec(`UPDATE stations SET status='offline' WHERE status='online' AND last_heartbeat < ?`,
		fmtTime(cutoff))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// UpsertSensor 写入/更新传感器配置。
func (s *Store) UpsertSensor(stationID string, kind model.SensorKind, warnAt, critAt float64, calibratedAt time.Time) error {
	_, err := s.db.Exec(`INSERT INTO sensors(id,station_id,kind,unit,warn_at,crit_at,calibrated_at)
		VALUES(?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET unit=excluded.unit, warn_at=excluded.warn_at, crit_at=excluded.crit_at, calibrated_at=excluded.calibrated_at`,
		stationID+":"+string(kind), stationID, string(kind), kind.Unit(), warnAt, critAt, fmtTime(calibratedAt))
	return err
}

// ListSensorsByStation 返回站上传感器配置。
func (s *Store) ListSensorsByStation(stationID string) ([]SensorConfig, error) {
	rows, err := s.db.Query(`SELECT id,station_id,kind,unit,warn_at,crit_at,calibrated_at FROM sensors WHERE station_id=? ORDER BY kind`, stationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SensorConfig
	for rows.Next() {
		var c SensorConfig
		var kind, calib string
		if err := rows.Scan(&c.ID, &c.StationID, &kind, &c.Unit, &c.WarnAt, &c.CritAt, &calib); err != nil {
			return nil, err
		}
		c.Kind = model.SensorKind(kind)
		c.CalibratedAt = parseTime(calib)
		out = append(out, c)
	}
	return out, rows.Err()
}

// SensorConfig 是传感器表行结构。
type SensorConfig struct {
	ID           string
	StationID    string
	Kind         model.SensorKind
	Unit         string
	WarnAt       float64
	CritAt       float64
	CalibratedAt time.Time
}
