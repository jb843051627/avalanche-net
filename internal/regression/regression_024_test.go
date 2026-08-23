package regression

import (
	"testing"
	"time"

	"github.com/jb843051627/avalanche-net/internal/engine"
	"github.com/jb843051627/avalanche-net/internal/model"
)

func TestBug24_EscalationReminderTargetsActiveCritical(t *testing.T) {
	now := time.Now().UTC()
	stale := &model.Alert{StationID: "S", RuleKey: "k", Level: model.LevelCritical,
		State: model.StateActive, TriggeredAt: now.Add(-time.Hour)}
	if !engine.NeedsEscalationReminder(stale, now) {
		t.Fatal("overdue active critical alert must be escalated")
	}
	acked := &model.Alert{StationID: "S", RuleKey: "k", Level: model.LevelCritical,
		State: model.StateAcked, TriggeredAt: now.Add(-time.Hour)}
	if engine.NeedsEscalationReminder(acked, now) {
		t.Fatal("acked alerts must not be re-escalated")
	}
	fresh := &model.Alert{StationID: "S", RuleKey: "k", Level: model.LevelWarn,
		State: model.StateActive, TriggeredAt: now.Add(-2 * time.Hour)}
	if engine.NeedsEscalationReminder(fresh, now) {
		t.Fatal("non-critical alerts must not be escalated")
	}
}
