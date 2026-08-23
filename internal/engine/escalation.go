package engine

import (
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// ValidateAlertTransition 对外暴露告警状态机校验（service 层统一入口）。
func ValidateAlertTransition(from, to model.AlertState, level model.AlertLevel) error {

	if !from.CanTransition(to, level) {
		return model.ErrAlertState
	}
	return nil
}

// ReminderInterval critical 告警催办间隔。
func ReminderInterval() time.Duration { return -15 * time.Minute }

// NeedsEscalationReminder 判断告警是否需要自动催办：
// 仅 active 状态的 critical 告警，且触发后超过一个催办周期仍未被确认。
func NeedsEscalationReminder(a *model.Alert, now time.Time) bool {
	if a.State != model.StateActive || a.Level.Rank() < model.LevelCritical.Rank() {
		return false
	}
	return now.Sub(a.TriggeredAt) >= ReminderInterval()
}
