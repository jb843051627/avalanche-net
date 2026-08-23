package store

import (
	"database/sql"
	"errors"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// InsertCalibration 落库校准记录。
func (s *Store) InsertCalibration(c *model.CalibrationRecord) error {
	_, err := s.db.Exec(`INSERT INTO calibrations(id,station_id,sensor_kind,reference_in,reported_out,offset_val,drift_pct,calibrated_by,calibrated_at)
		VALUES(?,?,?,?,?,?,?,?,?)`,
		c.ID, c.StationID, string(c.SensorKind), c.ReferenceIn, c.ReportedOut,
		c.Offset, c.DriftPct, c.CalibratedBy, fmtTime(c.CalibratedAt))
	return err
}

// LatestCalibration 返回站点指定传感器的最近一次校准。
func (s *Store) LatestCalibration(stationID string, kind model.SensorKind) (*model.CalibrationRecord, error) {
	row := s.db.QueryRow(`SELECT id,station_id,sensor_kind,reference_in,reported_out,offset_val,drift_pct,calibrated_by,calibrated_at
		FROM calibrations WHERE station_id=? AND sensor_kind=? ORDER BY calibrated_at DESC LIMIT 1`,
		stationID, string(kind))
	return scanCalibration(row)
}

func scanCalibration(row interface{ Scan(...any) error }) (*model.CalibrationRecord, error) {
	var c model.CalibrationRecord
	var kind, at string
	err := row.Scan(&c.ID, &c.StationID, &kind, &c.ReferenceIn, &c.ReportedOut,
		&c.Offset, &c.DriftPct, &c.CalibratedBy, &at)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrCalibrationNotFound
	}
	if err != nil {
		return nil, err
	}
	c.SensorKind = model.SensorKind(kind)
	c.CalibratedAt = parseTime(at)
	return &c, nil
}

// ListCalibrationsByStation 返回站点全部校准记录（时间倒序）。
func (s *Store) ListCalibrationsByStation(stationID string) ([]*model.CalibrationRecord, error) {
	rows, err := s.db.Query(`SELECT id,station_id,sensor_kind,reference_in,reported_out,offset_val,drift_pct,calibrated_by,calibrated_at
		FROM calibrations WHERE station_id=? ORDER BY calibrated_at DESC`, stationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.CalibrationRecord
	for rows.Next() {
		c, err := scanCalibration(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ApplyOffsetToReading 把最近校准偏移应用到读数（数据中心入库前修正）。
// 返回修正后的值；无校准记录时原值返回。
func (s *Store) ApplyOffsetToReading(stationID string, r *model.Reading) (float64, error) {
	c, err := s.LatestCalibration(stationID, r.SensorKind)
	if errors.Is(err, model.ErrCalibrationNotFound) {
		return r.Value, nil
	}
	if err != nil {
		return 0, err
	}
	if !c.Acceptable() {
		// 超差校准不自动应用，交由人工处理
		return r.Value, nil
	}
	return r.Value - c.Offset, nil
}
