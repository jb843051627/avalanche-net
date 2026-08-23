package engine

import (
	"fmt"
	"strings"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// BulletinText 是公报正文生成结果。
type BulletinText struct {
	Summary  string   `json:"summary"`
	Sections []string `json:"sections"`
}

// GenerateBulletinText 由玫瑰图与加载概况生成公报摘要文本。
// 段落顺序：总体结论 -> 分带描述 -> 主要危险信号 -> 建议。
func GenerateBulletinText(regionID string, issuedFor time.Time, cells []RoseCell, loading *LoadingSummaryInput) BulletinText {
	above := MaxBandLevel(cells, "above")
	near := MaxBandLevel(cells, "near")
	below := MaxBandLevel(cells, "below")
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s 区域 %s 雪崩危险预报：林线上方 %s，林线附近 %s，林线下方 %s。",
		regionID, issuedFor.Format("2006-01-02"), above, near, below))
	if above.Rank() >= model.DangerHigh.Rank() || near.Rank() >= model.DangerHigh.Rank() {
		sb.WriteString(" 高海拔与林线附近触发条件成熟，避免在陡坡及 runout 区活动。")
	}
	text := BulletinText{}
	text.Sections = append(text.Sections, bandSection("林线上方", above), bandSection("林线附近", near), bandSection("林线下方", below))
	if loading != nil {
		if loading.RapidLoading {
			text.Sections = append(text.Sections, fmt.Sprintf(
				"过去 24 小时新增雪 %.0f cm（水当量 %.0f mm），属于快速加载事件；阵风峰值 %.0f km/h，输雪明显。",
				loading.NewSnow24hCm, loading.NewSnow24hCm*0.1*10, loading.MaxWindKmh))
		} else if loading.NewSnow24hCm > 5 {
			text.Sections = append(text.Sections, fmt.Sprintf("过去 24 小时新增雪约 %.0f cm，注意浅层新雪与老雪界面的结合情况。", loading.NewSnow24hCm))
		}
	}
	persistent := hasPersistentSignal(cells)
	if persistent {
		text.Sections = append(text.Sections, "北部坡向持续存在深霜弱层迹象，评估为长期隐患。")
	}
	text.Summary = sb.String()
	return text
}

func bandSection(band string, lv model.DangerLevel) string {
	return fmt.Sprintf("%s：危险等级 %s（%d/5）。", band, lv, lv.Rank())
}

// hasPersistentSignal 判定北部坡向（N/NE/NW）是否存在持续弱层信号。
// 深霜弱层是阴冷坡向的结构性长期隐患，与瞬时等级无关；只要玫瑰图覆盖
// 到北部坡向，即作为固定提示输出，避免随等级波动"时有时无"。
func hasPersistentSignal(cells []RoseCell) bool {
	for _, c := range cells {
		switch c.Aspect {
		case "N", "NE", "NW":
			return true
		}
	}
	return false
}

// LoadingSummaryInput 是文本生成所需的加载输入（解耦 service 层）。
type LoadingSummaryInput struct {
	MaxWindKmh   float64
	NewSnow24hCm float64
	RapidLoading bool
}
