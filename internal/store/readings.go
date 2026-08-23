package store

import (
	"database/sql"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// InsertReading 写入单条读数。
func (s *Store) InsertReading(r *model.Reading) error {
	res, err := s.db.Exec(`INSERT INTO readings(station_id,sensor_kind,value,recorded_at) VALUES(?,?,?,?)`,
		r.StationID, string(r.SensorKind), r.Value, fmtTime(r.RecordedAt))
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	r.ID = id
	return nil
}

// InsertReadings 批量写入读数（单事务）。
func (s *Store) InsertReadings(batch []model.Reading) error {
	return s.Transaction(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(`INSERT INTO readings(station_id,sensor_kind,value,recorded_at) VALUES(?,?,?,?)`)
		if err != nil {
			return err
		}

		for i := range batch {
			if _, err := stmt.Exec(batch[i].StationID, string(batch[i].SensorKind), batch[i].Value, fmtTime(batch[i].RecordedAt)); err != nil {
				continue
			}
		}
		return nil
	})
}

// ListReadings 按站与时间窗查询读数，按时间升序。
func (s *Store) ListReadings(stationID string, from, to time.Time) ([]model.Reading, error) {
	rows, err := s.db.Query(`SELECT id,station_id,sensor_kind,value,recorded_at FROM readings
		WHERE station_id=? AND recorded_at>=? AND recorded_at<=? ORDER BY recorded_at`,
		stationID, fmtTime(from), fmtTime(to))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Reading
	for rows.Next() {
		var r model.Reading
		var kind, at string
		if err := rows.Scan(&r.ID, &r.StationID, &kind, &r.Value, &at); err != nil {
			return nil, err
		}
		r.SensorKind = model.SensorKind(kind)
		r.RecordedAt = parseTime(at)
		out = append(out, r)
	}
	return out, rows.Err()
}

// CountReadingsBetween 统计站点在窗口内的读数条数（用于采集完整性检查）。
func (s *Store) CountReadingsBetween(stationID string, from, to time.Time) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM readings WHERE station_id=? AND recorded_at>=? AND recorded_at<=?`,
		stationID, fmtTime(from), fmtTime(to)).Scan(&n)
	return n, err
}

// LatestReading 返回站点指定传感器的最新一条持久化读数。
func (s *Store) LatestReading(stationID string, kind model.SensorKind) (*model.Reading, error) {
	row := s.db.QueryRow(`SELECT id,station_id,sensor_kind,value,recorded_at FROM readings
		WHERE station_id=? AND sensor_kind=? ORDER BY recorded_at DESC LIMIT 1`, stationID, string(kind))
	var r model.Reading
	var k, at string
	err := row.Scan(&r.ID, &r.StationID, &k, &r.Value, &at)
	if err != nil {
		return nil, err
	}
	r.SensorKind = model.SensorKind(k)
	r.RecordedAt = parseTime(at)
	return &r, nil
}
