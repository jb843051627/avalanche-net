package model

import (
	"errors"
	"time"
)

var (
	ErrAlertNotFound = errors.New("alert not found")
	ErrAlertState    = errors.New("invalid alert state transition")
	ErrDeduplicated  = errors.New("duplicate alert inside dedup window")
)

// AlertLevel 告警级别。
type AlertLevel string

const (
	LevelInfo     AlertLevel = "info"
	LevelWarn     AlertLevel = "warn"
	LevelCritical AlertLevel = "critical"
)

// Rank 返回告警级别的数值权重，用于升级比较。
func (l AlertLevel) Rank() int {
	switch l {
	case LevelInfo:
		return 1
	case LevelWarn:
		return 2
	case LevelCritical:
		return 3
	}
	return 0
}

// AlertState 告警状态机：active -> acked -> resolved；active 可直接 resolved。
type AlertState string

const (
	StateActive   AlertState = "active"
	StateAcked    AlertState = "acked"
	StateResolved AlertState = "resolved"
)

// CanTransition 校验告警状态迁移合法性。
// 规则：resolved 必须先 acked（critical 不允许跳过确认直接关闭）。
func (s AlertState) CanTransition(to AlertState, level AlertLevel) bool {
	switch s {
	case StateActive:
		if to == StateAcked {
			return true
		}
		if to == StateResolved {
			return level.Rank() < LevelCritical.Rank()
		}
	case StateAcked:
		if to == StateResolved {
			return true
		}
	}
	return false
}

// Alert 是一条告警记录。
type Alert struct {
	ID          string     `json:"id"`
	StationID   string     `json:"station_id"`
	RuleKey     string     `json:"rule_key"`
	Level       AlertLevel `json:"level"`
	State       AlertState `json:"state"`
	Reason      string     `json:"reason"`
	Value       float64    `json:"value"`
	TriggeredAt time.Time  `json:"triggered_at"`
	AckedBy     string     `json:"acked_by,omitempty"`
	AckedAt     *time.Time `json:"acked_at,omitempty"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// DedupKey 生成去重键：同站同规则在窗口内视为重复。
func (a *Alert) DedupKey() string {
	return a.StationID + "|" + a.RuleKey
}
