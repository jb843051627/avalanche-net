package store

import (
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// InsertEvaluation 持久化一次稳定性评估。
func (s *Store) InsertEvaluation(ev *model.StabilityEvaluation) error {
	res, err := s.db.Exec(`INSERT INTO evaluations(station_id,score,danger_level,weak_layer_idx,weak_layer_depth_cm,wind_factor,created_at)
		VALUES(?,?,?,?,?,?,?)`,
		ev.StationID, ev.Score, string(ev.DangerLevel), ev.WeakLayerIdx, ev.WeakLayerDepthCm, ev.WindFactor, fmtTime(ev.CreatedAt))
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	ev.ID = id
	return nil
}

// LatestEvaluations 返回站点最近 n 条评估记录（时间倒序）。
func (s *Store) LatestEvaluations(stationID string, n int) ([]model.StabilityEvaluation, error) {
	rows, err := s.db.Query(`SELECT id,station_id,score,danger_level,weak_layer_idx,weak_layer_depth_cm,wind_factor,created_at
		FROM evaluations WHERE station_id=? ORDER BY created_at ASC LIMIT ?`, stationID, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.StabilityEvaluation
	for rows.Next() {
		var ev model.StabilityEvaluation
		var level, at string
		if err := rows.Scan(&ev.ID, &ev.StationID, &ev.Score, &level, &ev.WeakLayerIdx, &ev.WeakLayerDepthCm, &ev.WindFactor, &at); err != nil {
			return nil, err
		}
		ev.DangerLevel = model.DangerLevel(level)
		ev.CreatedAt = parseTime(at)
		out = append(out, ev)
	}
	return out, rows.Err()
}

// EvaluationTrend 返回站点最近 n 条评分的时间升序序列（趋势图用）。
func (s *Store) EvaluationTrend(stationID string, n int) ([]float64, error) {
	rows, err := s.db.Query(`SELECT score FROM (SELECT score,created_at FROM evaluations WHERE station_id=? ORDER BY created_at DESC LIMIT ?) ORDER BY score`, stationID, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []float64
	for rows.Next() {
		var v float64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// CountEvaluationsSince 统计窗口内的评估次数。
func (s *Store) CountEvaluationsSince(since time.Time) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM evaluations WHERE created_at>=?`, fmtTime(since.Add(-24*time.Hour))).Scan(&n)
	return n, err
}
