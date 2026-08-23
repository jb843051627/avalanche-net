package service

import (
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// ScheduleInspection 创建巡检任务。
func (s *Service) ScheduleInspection(stationID string, due time.Time, assignee string) (*model.InspectionTask, error) {
	if _, err := s.store.GetStation(stationID); err != nil {
		return nil, err
	}
	t := &model.InspectionTask{
		ID:        "insp-" + stationID + "-" + due.UTC().Format("20060102"),
		StationID: stationID,
		DueDate:   due.UTC(),
		Status:    model.InspectionDue,
		Assignee:  assignee,
		CreatedAt: s.clk.Now().UTC(),
	}
	if err := s.store.InsertInspection(t); err != nil {
		return nil, err
	}
	s.met.Inc("inspection.scheduled")
	return t, nil
}

// StartInspection 把任务推进到 in_progress。
func (s *Service) StartInspection(id string) (*model.InspectionTask, error) {
	t, err := s.store.GetInspection(id)
	if err != nil {
		return nil, err
	}
	if !t.Status.CanTransition(model.InspectionInProgress) {
		return nil, model.ErrInspectionState
	}
	t.Status = model.InspectionInProgress
	if err := s.store.UpdateInspection(t); err != nil {
		return nil, err
	}
	return t, nil
}

// CompleteInspection 完成巡检：必须携带 completion notes，写完成时间。
func (s *Service) CompleteInspection(id, notes string) (*model.InspectionTask, error) {
	t, err := s.store.GetInspection(id)
	if err != nil {
		return nil, err
	}
	if notes == "" {
		return nil, model.ErrNotesRequired
	}
	if !t.Status.CanTransition(model.InspectionCompleted) {
		return nil, model.ErrInspectionState
	}
	now := s.clk.Now().UTC()
	t.Status = model.InspectionCompleted
	t.Notes = notes
	t.CompletedAt = &now
	if err := s.store.UpdateInspection(t); err != nil {
		return nil, err
	}
	s.met.Inc("inspection.completed")
	return t, nil
}

// CancelInspection 取消任务（due/in_progress 均可取消）。
func (s *Service) CancelInspection(id string) (*model.InspectionTask, error) {
	t, err := s.store.GetInspection(id)
	if err != nil {
		return nil, err
	}
	if !t.Status.CanTransition(model.InspectionCancelled) {
		return nil, model.ErrInspectionState
	}
	t.Status = model.InspectionCancelled
	if err := s.store.UpdateInspection(t); err != nil {
		return nil, err
	}
	return t, nil
}

// ListInspections 查询巡检任务列表。
func (s *Service) ListInspections(stationID string, status model.InspectionStatus) ([]*model.InspectionTask, error) {
	return s.store.ListInspections(stationID, status)
}

// OverdueCount 统计逾期未完成巡检数。
func (s *Service) OverdueCount() (int, error) {
	return s.store.CountOverdueInspections(s.clk.Now().UTC())
}
