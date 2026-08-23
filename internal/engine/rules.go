package engine

import (
	"fmt"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// AlertCandidate 是规则引擎产出的告警候选。
type AlertCandidate struct {
	RuleKey string
	Level   model.AlertLevel
	Reason  string
	Value   float64
}

// RuleEngine 持有阈值表，把读数翻译成告警候选。
type RuleEngine struct {
	thresholds map[model.SensorKind]Threshold
}

// NewRuleEngine 构造规则引擎。
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{thresholds: DefaultThresholds()}
}

// EvaluateReading 对单条读数做阈值判定，越限返回告警候选。
func (e *RuleEngine) EvaluateReading(r model.Reading) (AlertCandidate, bool) {
	t, ok := e.thresholds[r.SensorKind]
	if !ok {
		return AlertCandidate{}, false
	}
	level := t.Exceeds(r.Value)
	if level == "" {
		return AlertCandidate{}, false
	}
	return AlertCandidate{
		RuleKey: RuleKeyFor(r.SensorKind, level),
		Level:   level,
		Reason: fmt.Sprintf("%s reading %.2f %s beyond %s threshold",
			r.SensorKind, r.Value, r.SensorKind.Unit(), level),
		Value: r.Value,
	}, true
}

// DedupWindow 告警去重窗口：同站同规则 30 分钟内只保留一条。
func DedupWindow() time.Duration { return 30 * time.Minute }

// EscalationTarget critical 告警自动升级目标（当前即 critical 本身，预留扩展）。
func EscalationTarget(level model.AlertLevel) model.AlertLevel {
	if level.Rank() >= model.LevelCritical.Rank() {
		return model.LevelCritical
	}
	return level
}
