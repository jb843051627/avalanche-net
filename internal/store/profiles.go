package store

import (
	"database/sql"
	"errors"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// InsertProfile 落库剖面及全部雪层（单事务）。
func (s *Store) InsertProfile(p *model.SnowProfile) error {
	return s.Transaction(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`INSERT INTO snow_profiles(id,station_id,observed_at,observer,total_cm) VALUES(?,?,?,?,?)`,
			p.ID, p.StationID, fmtTime(p.ObservedAt), p.Observer, p.TotalCm); err != nil {
			return err
		}
		stmt, err := tx.Prepare(`INSERT INTO snow_layers(profile_id,idx,depth_from_cm,depth_to_cm,density_kgm3,grain_shape,hardness,temp_c)
			VALUES(?,?,?,?,?,?,?,?)`)
		if err != nil {
			return err
		}
		defer stmt.Close()
		for _, l := range p.Layers {
			if _, err := stmt.Exec(p.ID, l.Index, l.DepthFromCm, l.DepthToCm, l.DensityKgM3,
				string(l.GrainShape), string(l.Hardness), l.TempC); err != nil {
				return err
			}
		}
		return nil
	})
}

// GetProfile 加载剖面（含雪层，按 idx 升序）；未找到返回 ErrProfileNotFound。
func (s *Store) GetProfile(id string) (*model.SnowProfile, error) {
	row := s.db.QueryRow(`SELECT id,station_id,observed_at,observer,total_cm FROM snow_profiles WHERE id=?`, id)
	var p model.SnowProfile
	var at string
	err := row.Scan(&p.ID, &p.StationID, &at, &p.Observer, &p.TotalCm)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrProfileNotFound
	}
	if err != nil {
		return nil, err
	}
	p.ObservedAt = parseTime(at)
	layers, err := s.loadLayers(p.ID)
	if err != nil {
		return nil, err
	}
	p.Layers = layers
	return &p, nil
}

// LatestProfile 返回站点最近一次观测的剖面。
func (s *Store) LatestProfile(stationID string) (*model.SnowProfile, error) {
	var id string
	err := s.db.QueryRow(`SELECT id FROM snow_profiles WHERE station_id=? ORDER BY observed_at DESC LIMIT 1`, stationID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrProfileNotFound
	}
	if err != nil {
		return nil, err
	}
	return s.GetProfile(id)
}

func (s *Store) loadLayers(profileID string) ([]model.SnowLayer, error) {
	rows, err := s.db.Query(`SELECT idx,depth_from_cm,depth_to_cm,density_kgm3,grain_shape,hardness,temp_c
		FROM snow_layers WHERE profile_id=? ORDER BY idx`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.SnowLayer
	for rows.Next() {
		var l model.SnowLayer
		var grain, hardness string
		if err := rows.Scan(&l.Index, &l.DepthFromCm, &l.DepthToCm, &l.DensityKgM3, &grain, &hardness, &l.TempC); err != nil {
			return nil, err
		}
		l.GrainShape = model.GrainShape(grain)
		l.Hardness = model.Hardness(hardness)
		out = append(out, l)
	}
	return out, rows.Err()
}

// CountProfilesByStation 统计站点历史剖面数量。
func (s *Store) CountProfilesByStation(stationID string) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM snow_profiles WHERE station_id=?`, stationID).Scan(&n)
	return n, err
}
