package engine

import (
	"github.com/jb843051627/avalanche-net/internal/model"
)

// ProfileDiff 是两次剖面的演化对比结果。
type ProfileDiff struct {
	NewSnowCm      float64              `json:"new_snow_cm"`   // 新增积雪（表面下探深度差）
	SettlementCm   float64              `json:"settlement_cm"` // 沉降（总深收缩）
	SettlementPct  float64              `json:"settlement_pct"`
	LayerCountFrom int                  `json:"layer_count_from"`
	LayerCountTo   int                  `json:"layer_count_to"`
	HardeningIdx   []int                `json:"hardening_idx"` // 变硬的层序号（以新剖面计）
	WeakeningIdx   []int                `json:"weakening_idx"` // 变软的层序号
	NewWeakLayers  []WeakLayerCandidate `json:"new_weak_layers"`
	PersistentKept bool                 `json:"persistent_kept"` // 原弱层是否仍存在
}

// CompareProfiles 对比旧、新两次剖面。约定 older.ObservedAt <= newer.ObservedAt。
// 表面新增 = newer.TotalCm - older.TotalCm（>0 为新雪，<0 为沉降/消融）。
func CompareProfiles(older, newer *model.SnowProfile) ProfileDiff {
	diff := ProfileDiff{
		LayerCountFrom: len(older.Layers),
		LayerCountTo:   len(newer.Layers),
	}
	delta := newer.TotalCm - older.TotalCm
	if delta > 0 {
		diff.NewSnowCm = delta
	} else {
		diff.SettlementCm = -delta
		if older.TotalCm > 0 {
			diff.SettlementPct = -delta / older.TotalCm * 100
		}
	}
	oldByDepth := matchByDepth(older.Layers)
	for i, l := range newer.Layers {
		prev, ok := oldByDepth(nearestBucket(l.DepthToCm))
		if !ok {
			continue
		}
		switch {
		case l.Hardness.Score() > prev.Hardness.Score()+1:
			diff.HardeningIdx = append(diff.HardeningIdx, i+1)
		case l.Hardness.Score() < prev.Hardness.Score()-1:
			diff.WeakeningIdx = append(diff.WeakeningIdx, i+1)
		}
	}
	newWeak, hasNew := FindWeakLayer(newer.Layers)
	if hasNew {
		diff.NewWeakLayers = append(diff.NewWeakLayers, newWeak)
	}
	diff.PersistentKept = weakLayerPersists(older, newer)
	return diff
}

// nearestBucket 把深度按 5cm 分桶，吸收观测误差。
func nearestBucket(depth float64) float64 {
	return float64(int(depth/5+0.5)) * 5
}

func matchByDepth(layers []model.SnowLayer) func(float64) (model.SnowLayer, bool) {
	index := make(map[float64]model.SnowLayer, len(layers))
	for _, l := range layers {
		index[nearestBucket(l.DepthToCm)] = l
	}
	return func(d float64) (model.SnowLayer, bool) {
		l, ok := index[d]
		return l, ok
	}
}

// weakLayerPersisters 判断旧剖面的最深弱层在新剖面同埋深处是否仍为弱层。
func weakLayerPersists(older, newer *model.SnowProfile) bool {
	oldWeak, found := FindWeakLayer(older.Layers)
	if !found {
		return false
	}
	for _, l := range newer.Layers {
		if absF(l.DepthToCm-oldWeak.DepthCm) < 10 && WeakLayerScore(l, l.TempC) >= 50 {
			return true
		}
	}
	return false
}

func absF(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
