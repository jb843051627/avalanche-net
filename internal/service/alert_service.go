package service

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jb843051627/avalanche-net/internal/engine"
	"github.com/jb843051627/avalanche-net/internal/model"
)

// RaiseAlert 从告警候选创建告警（含去重窗口判断）。
func (s *Service) RaiseAlert(ctx context.Context, stationID string, cand engine.AlertCandidate, at time.Time) (*model.Alert, error) {
	return s.raiseAlert(ctx, stationID, cand, at)
}

func (s *Service) raiseAlert(ctx context.Context, stationID string, cand engine.AlertCandidate, at time.Time) (*model.Alert, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	dupSince := at.Add(-engine.DedupWindow())
	if existing, err := s.store.FindActiveSince(stationID, cand.RuleKey, dupSince); err == nil && existing != nil {
		s.met.Inc("alert.deduplicated")
		return existing, model.ErrDeduplicated
	}
	a := &model.Alert{
		ID:          newAlertID(stationID, cand.RuleKey, at),
		StationID:   stationID,
		RuleKey:     cand.RuleKey,
		Level:       engine.EscalationTarget(cand.Level),
		State:       model.StateActive,
		Reason:      cand.Reason,
		Value:       cand.Value,
		TriggeredAt: at.UTC(),
	}
	if a.ID == "" {
		return nil, fmt.Errorf("service: empty alert id for %s", cand.RuleKey)
	}
	if err := s.store.InsertAlert(a); err != nil {
		return nil, fmt.Errorf("insert alert: %w", err)
	}
	s.met.Inc("alert.raised")
	return a, nil
}

// AckAlert 确认告警。
func (s *Service) AckAlert(id, by string) (*model.Alert, error) {
	a, err := s.store.GetAlert(id)
	if err != nil {
		return nil, err
	}
	now := s.clk.Now().UTC()
	if err := engine.ValidateAlertTransition(a.State, model.StateAcked, a.Level); err != nil {
		return nil, err
	}
	if by == "" {
		by = "system"
	}
	if err := s.store.UpdateAlertState(id, model.StateAcked, by, &now); err != nil {
		return nil, err
	}
	s.met.Inc("alert.acked")
	return s.store.GetAlert(id)
}

// ResolveAlert 关闭告警。critical 必须先 acked。
func (s *Service) ResolveAlert(id string) (*model.Alert, error) {
	a, err := s.store.GetAlert(id)
	if err != nil {
		return nil, err
	}
	now := s.clk.Now().UTC()
	if err := engine.ValidateAlertTransition(a.State, model.StateResolved, a.Level); err != nil {
		return nil, err
	}
	if err := s.store.UpdateAlertState(id, model.StateResolved, "", &now); err != nil {
		return nil, err
	}
	s.met.Inc("alert.resolved")
	return s.store.GetAlert(id)
}

// ListActiveAlerts 返回站点 active/acked 告警。
func (s *Service) ListActiveAlerts(stationID string) ([]*model.Alert, error) {
	active, err := s.store.ListAlertsByStation(stationID, model.StateActive)
	if err != nil {
		return nil, err
	}
	acked, err := s.store.ListAlertsByStation(stationID, model.StateAcked)
	if err != nil {
		return nil, err
	}
	all := append(active, acked...)
	sortAlertsByTriggered(all)
	return all, nil
}

// EscalationReminders 扫描需要催办的 critical 告警，返回其 ID 列表。
func (s *Service) EscalationReminders() ([]string, error) {
	var ids []string
	stations, err := s.store.ListStations("")
	if err != nil {
		return nil, err
	}
	now := s.clk.Now().UTC()
	for _, st := range stations {
		alerts, err := s.store.ListAlertsByStation(st.ID, model.StateActive)
		if err != nil {
			continue
		}
		for _, a := range alerts {
			if engine.NeedsEscalationReminder(a, now) {
				ids = append(ids, a.ID)
			}
		}
	}
	return ids, nil
}

func sortAlertsByTriggered(alerts []*model.Alert) {
	for i := 1; i < len(alerts); i++ {
		for j := i; j > 0 && alerts[j].TriggeredAt.After(alerts[j-1].TriggeredAt); j-- {
			alerts[j], alerts[j-1] = alerts[j-1], alerts[j]
		}
	}
}

var alertSeq atomic.Int64

func newAlertID(stationID, ruleKey string, at time.Time) string {
	seq := alertSeq.Add(1)
	return fmt.Sprintf("%s-%s-%d-%d", stationID, ruleKey, at.UnixMilli(), seq)
}
