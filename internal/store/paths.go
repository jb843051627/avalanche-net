package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// InsertPath 落库雪崩路径。
func (s *Store) InsertPath(p *model.AvalanchePath) error {
	var prevEvt any
	if p.LastEventAt != nil {
		prevEvt = fmtTime(*p.LastEventAt)
	}
	_, err := s.db.Exec(`INSERT INTO avalanche_paths(id,region_id,name,start_elev_m,end_elev_m,aspect,slope_deg,length_m,hits_road,hits_structs,last_event_at,event_count_12m,registered_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		p.ID, p.RegionID, p.Name, p.StartElevM, p.EndElevM, string(p.Aspect), p.SlopeDeg,
		p.LengthM, p.HitsRoad, p.HitsStructs, prevEvt, p.EventCount12m, fmtTime(p.RegisteredAt))
	return err
}

const pathCols = `id,region_id,name,start_elev_m,end_elev_m,aspect,slope_deg,length_m,hits_road,hits_structs,last_event_at,event_count_12m,registered_at`

func scanPath(row interface{ Scan(...any) error }) (*model.AvalanchePath, error) {
	var p model.AvalanchePath
	var aspect, registered string
	var prevEvt sql.NullString
	var road, structs int
	err := row.Scan(&p.ID, &p.RegionID, &p.Name, &p.StartElevM, &p.EndElevM, &aspect,
		&p.SlopeDeg, &p.LengthM, &road, &structs, &prevEvt, &p.EventCount12m, &registered)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrPathNotFound
	}
	if err != nil {
		return nil, err
	}
	p.Aspect = model.Aspect(aspect)
	p.HitsRoad = road == 1
	p.HitsStructs = structs == 1
	p.RegisteredAt = parseTime(registered)
	if prevEvt.Valid && prevEvt.String != "" {
		t := parseTime(prevEvt.String)
		p.LastEventAt = &t
	}
	return &p, nil
}

// GetPath 按 ID 查询路径。
func (s *Store) GetPath(id string) (*model.AvalanchePath, error) {
	return scanPath(s.db.QueryRow(`SELECT `+pathCols+` FROM avalanche_paths WHERE id=?`, id))
}

// ListPathsByRegion 返回区域全部登记路径（按名称排序）。
func (s *Store) ListPathsByRegion(regionID string) ([]*model.AvalanchePath, error) {
	rows, err := s.db.Query(`SELECT `+pathCols+` FROM avalanche_paths WHERE region_id=? ORDER BY name`, regionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.AvalanchePath
	for rows.Next() {
		p, err := scanPath(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// MarkPathEvent 更新路径最近事件时间并累加近一年事件计数。
func (s *Store) MarkPathEvent(id string, at time.Time) error {
	res, err := s.db.Exec(`UPDATE avalanche_paths SET last_event_at=?, event_count_12m=event_count_12m+1 WHERE id=?`,
		fmtTime(at), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrPathNotFound
	}
	return nil
}
