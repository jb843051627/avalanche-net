package engine

import (
	"fmt"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// RoseCell 是危险玫瑰图的一个坡向×海拔带单元。
type RoseCell struct {
	Aspect       string            `json:"aspect"`
	Band         string            `json:"band"` // above/near/below
	DangerLevel  model.DangerLevel `json:"danger_level"`
	StationCount int               `json:"station_count"`
}

// BuildRose 把站点最新评估聚合到 8 方位 × 3 海拔带的玫瑰图。
// 同一单元取最高等级。
func BuildRose(entries []RoseEntry) []RoseCell {
	cells := make(map[string]*RoseCell, 24)
	for _, e := range entries {
		band := ElevationBand(e.ElevationM)
		key := fmt.Sprintf("%s|%s", e.Aspect, band)
		cell := cells[key]
		if cell == nil {
			cell = &RoseCell{Aspect: string(e.Aspect), Band: band, DangerLevel: model.DangerLow}
			cells[key] = cell
		}
		if e.Level.Rank() > cell.DangerLevel.Rank() {
			cell.DangerLevel = e.Level
		}
		cell.StationCount++
	}
	out := make([]RoseCell, 0, len(cells))
	for _, c := range cells {
		out = append(out, *c)
	}
	sortCells(out)
	return out
}

// RoseEntry 是玫瑰图输入行。
type RoseEntry struct {
	Aspect     model.Aspect
	ElevationM float64
	Level      model.DangerLevel
}

func sortCells(cells []RoseCell) {
	for i := 1; i < len(cells); i++ {
		for j := i; j > 0 && cellLess(cells[j], cells[j-1]); j-- {
			cells[j], cells[j-1] = cells[j-1], cells[j]
		}
	}
}

func cellLess(a, b RoseCell) bool {
	if a.Aspect != b.Aspect {
		return a.Aspect < b.Aspect
	}
	return a.Band < b.Band
}

// MaxBandLevel 返回玫瑰图中指定海拔带的最高等级（公报建议用）。
func MaxBandLevel(cells []RoseCell, band string) model.DangerLevel {
	max := model.DangerLow
	for _, c := range cells {
		if c.Band == band && c.DangerLevel.Rank() > max.Rank() {
			max = c.DangerLevel
		}
	}
	return max
}
