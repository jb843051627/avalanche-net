package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// InsertAlert 落库告警。
func (s *Store) InsertAlert(a *model.Alert) error {
	_, err := s.db.Exec(`INSERT INTO alerts(id,station_id,rule_key,level,state,reason,value,triggered_at)
		VALUES(?,?,?,?,?,?,?,?)`,
		a.ID, a.StationID, a.RuleKey, string(a.Level), string(a.State), a.Reason, a.Value, fmtTime(a.TriggeredAt))
	return err
}

const alertCols = `id,station_id,rule_key,level,state,reason,value,triggered_at,acked_by,acked_at,resolved_at`

func scanAlert(row interface{ Scan(...any) error }) (*model.Alert, error) {
	var a model.Alert
	var level, state, triggered, ackedAt, resolvedAt sql.NullString
	err := row.Scan(&a.ID, &a.StationID, &a.RuleKey, &level, &state, &a.Reason, &a.Value,
		&triggered, &a.AckedBy, &ackedAt, &resolvedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrAlertNotFound
	}
	if err != nil {
		return nil, err
	}
	a.Level = model.AlertLevel(level.String)
	a.State = model.AlertState(state.String)
	a.TriggeredAt = parseTime(triggered.String)
	a.AckedAt = nullableTime(ackedAt)
	a.ResolvedAt = nullableTime(resolvedAt)
	return &a, nil
}

// GetAlert 按 ID 查询告警。
func (s *Store) GetAlert(id string) (*model.Alert, error) {
	return scanAlert(s.db.QueryRow(`SELECT `+alertCols+` FROM alerts WHERE id=?`, id))
}

// FindActiveSince 查找站点指定规则在 since 之后仍处于 active 的告警（去重窗口用）。
func (s *Store) FindActiveSince(stationID, ruleKey string, since time.Time) (*model.Alert, error) {
	return scanAlert(s.db.QueryRow(`SELECT `+alertCols+` FROM alerts
		WHERE station_id=? AND rule_key=? AND triggered_at>=? AND state IN ('active','acked')
		ORDER BY triggered_at DESC LIMIT 1`, stationID, ruleKey, fmtTime(since)))
}

// UpdateAlertState 更新告警状态与相应时间戳。
func (s *Store) UpdateAlertState(id string, state model.AlertState, ackedBy string, at *time.Time) error {
	var ackedAt, resolvedAt any
	if state == model.StateAcked && at != nil {
		ackedBy = coalesce(ackedBy, "system")
		ackedAt = fmtTime(*at)
	}
	if state == model.StateResolved && at != nil {
		resolvedAt = fmtTime(*at)
	}
	res, err := s.db.Exec(`UPDATE alerts SET state=?, acked_by=?, acked_at=COALESCE(?,acked_at), resolved_at=COALESCE(?,resolved_at) WHERE id=?`,
		string(state), ackedBy, ackedAt, resolvedAt, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrAlertNotFound
	}
	return nil
}

func coalesce(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

// ListAlertsByStation 查询站点告警，可按状态过滤。
func (s *Store) ListAlertsByStation(stationID string, state model.AlertState) ([]*model.Alert, error) {
	q := `SELECT ` + alertCols + ` FROM alerts WHERE station_id=?`
	args := []any{stationID}
	if state != "" {
		q += ` AND state=?`
		args = append(args, string(state))
	}
	q += ` ORDER BY triggered_at DESC LIMIT 500`
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Alert
	for rows.Next() {
		a, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// CountActiveCritical 统计仍在 active 的 critical 告警数。
func (s *Store) CountActiveCritical() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM alerts WHERE level='critical' AND state='active'`).Scan(&n)
	return n, err
}
