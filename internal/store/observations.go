package store

import (
	"database/sql"
	"errors"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// InsertObservation 落库雪崩目击登记。
func (s *Store) InsertObservation(o *model.Observation) error {
	_, err := s.db.Exec(`INSERT INTO observations(id,region_id,station_id,observed_at,aspect,elevation_m,type,size,trigger,reporter,comment)
		VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		o.ID, o.RegionID, o.StationID, fmtTime(o.ObservedAt), string(o.Aspect), o.ElevationM,
		string(o.Type), int(o.Size), string(o.Trigger), o.Reporter, o.Comment)
	return err
}

const observationCols = `id,region_id,station_id,observed_at,aspect,elevation_m,type,size,trigger,reporter,comment`

func scanObservation(row interface{ Scan(...any) error }) (*model.Observation, error) {
	var o model.Observation
	var aspect, at, otype, trigger string
	var size int
	err := row.Scan(&o.ID, &o.RegionID, &o.StationID, &at, &aspect, &o.ElevationM,
		&otype, &size, &trigger, &o.Reporter, &o.Comment)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrObservationNotFound
	}
	if err != nil {
		return nil, err
	}
	o.Aspect = model.Aspect(aspect)
	o.ObservedAt = parseTime(at)
	o.Type = model.AvalancheType(otype)
	o.Size = model.AvalancheSize(size)
	o.Trigger = model.TriggerType(trigger)
	return &o, nil
}

// GetObservation 按 ID 查询。
func (s *Store) GetObservation(id string) (*model.Observation, error) {
	return scanObservation(s.db.QueryRow(`SELECT `+observationCols+` FROM observations WHERE id=?`, id))
}

// ListObservationsByRegion 返回区域目击记录（时间倒序，限 200 条）。
func (s *Store) ListObservationsByRegion(regionID string) ([]*model.Observation, error) {
	rows, err := s.db.Query(`SELECT `+observationCols+` FROM observations WHERE region_id=? ORDER BY observed_at DESC LIMIT 200`, regionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Observation
	for rows.Next() {
		o, err := scanObservation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// CountObservationsByTrigger 统计区域内各触发方式的目击数。
func (s *Store) CountObservationsByTrigger(regionID string) (map[string]int, error) {
	rows, err := s.db.Query(`SELECT trigger, COUNT(*) FROM observations WHERE region_id=? GROUP BY trigger`, regionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var k string
		var n int
		if err := rows.Scan(&k, &n); err != nil {
			return nil, err
		}
		out[k] = n
	}
	return out, rows.Err()
}
