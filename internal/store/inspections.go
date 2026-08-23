package store

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// InsertInspection 创建巡检任务。
func (s *Store) InsertInspection(t *model.InspectionTask) error {
	_, err := s.db.Exec(`INSERT INTO inspections(id,station_id,due_date,status,assignee,notes,created_at)
		VALUES(?,?,?,?,?,?,?)`,
		t.ID, t.StationID, fmtTime(t.DueDate), string(t.Status), t.Assignee, t.Notes, fmtTime(t.CreatedAt))
	if err != nil && strings.Contains(err.Error(), "UNIQUE") {
		return model.ErrStationExists
	}
	return err
}

const inspectionCols = `id,station_id,due_date,status,assignee,notes,created_at,completed_at`

func scanInspection(row interface{ Scan(...any) error }) (*model.InspectionTask, error) {
	var t model.InspectionTask
	var status, due, created string
	var completed sql.NullString
	err := row.Scan(&t.ID, &t.StationID, &due, &status, &t.Assignee, &t.Notes, &created, &completed)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrInspectionNotFound
	}
	if err != nil {
		return nil, err
	}
	t.Status = model.InspectionStatus(status)
	t.DueDate = parseTime(due)
	t.CreatedAt = parseTime(created)
	t.CompletedAt = nullableTime(completed)
	return &t, nil
}

// GetInspection 按 ID 查询巡检任务。
func (s *Store) GetInspection(id string) (*model.InspectionTask, error) {
	return scanInspection(s.db.QueryRow(`SELECT `+inspectionCols+` FROM inspections WHERE id=?`, id))
}

// UpdateInspection 更新状态/备注/完成时间。
func (s *Store) UpdateInspection(t *model.InspectionTask) error {
	var completed any
	if t.CompletedAt != nil {
		completed = fmtTime(*t.CompletedAt)
	}
	res, err := s.db.Exec(`UPDATE inspections SET status=?, assignee=?, notes=?, completed_at=? WHERE id=?`,
		string(t.Status), t.Assignee, t.Notes, completed, t.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrInspectionNotFound
	}
	return nil
}

// ListInspections 查询巡检任务，可按站点与状态过滤。
func (s *Store) ListInspections(stationID string, status model.InspectionStatus) ([]*model.InspectionTask, error) {
	q := `SELECT ` + inspectionCols + ` FROM inspections WHERE 1=1`
	args := []any{}
	if stationID != "" {
		q += ` AND station_id=?`
		args = append(args, stationID)
	}
	if status != "" {
		q += ` AND status=?`
		args = append(args, string(status))
	}
	q += ` ORDER BY due_date`
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.InspectionTask
	for rows.Next() {
		t, err := scanInspection(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// CountOverdueInspections 统计逾期未完成任务数。
func (s *Store) CountOverdueInspections(now time.Time) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM inspections WHERE status IN ('due','in_progress') AND due_date<?`,
		fmtTime(now)).Scan(&n)
	return n, err
}
