package engine

import (
	"github.com/jb843051627/avalanche-net/internal/model"
)

// WeakLayerCandidate 是剖面弱层候选的评分明细。
type WeakLayerCandidate struct {
	Index   int
	DepthCm float64
	Score   float64
}

// SlabScore 估算弱层上覆板层（slab）的驱动势：
// 上覆层越厚、越硬，释放时势能越大。
func SlabScore(layers []model.SnowLayer, weakIdx int) float64 {
	if weakIdx <= 0 || weakIdx > len(layers) {
		return 0
	}
	thickness := layers[weakIdx-1].DepthFromCm // 弱层埋深 = 上覆厚度
	hardSum := 0.0
	count := 0.0
	for i := 0; i < weakIdx-1; i++ {
		hardSum += float64(layers[i].Hardness.Score())
		count++
	}
	avgHard := 3.0
	if count > 0 {
		avgHard = hardSum / count
	}
	depthFactor := thickness / 100.0 // 米化
	if depthFactor > 2.0 {
		depthFactor = 2.0
	}
	return (depthFactor*30 + avgHard*5) / 90 * 40
}

// WeakLayerScore 对单层打分：晶型权重 × 软硬度 × 温度梯度贡献。
func WeakLayerScore(l model.SnowLayer, prevTemp float64) float64 {
	score := l.GrainShape.WeakLayerWeight() * 60
	hardnessBonus := float64(7-l.Hardness.Score()) * 4 // 越软越危险
	grad := l.TempGradient(prevTemp)
	gradBonus := grad
	if gradBonus > 20 {
		gradBonus = 20
	}
	total := score + hardnessBonus + gradBonus
	if total > 100 {
		total = 100
	}
	return total
}

// FindWeakLayer 扫描全部雪层，返回得分最高的弱层候选（>=50 分才算弱层）。
func FindWeakLayer(layers []model.SnowLayer) (WeakLayerCandidate, bool) {
	var best WeakLayerCandidate
	found := false
	prevTemp := 0.0
	for i, l := range layers {
		s := WeakLayerScore(l, prevTemp)
		prevTemp = l.TempC
		if s >= 50 && (!found || s > best.Score) {
			best = WeakLayerCandidate{Index: i + 1, DepthCm: l.DepthToCm, Score: s}
			found = true
		}
	}
	return best, found
}

// ProfileStability 综合剖面稳定性评分（0-100，越高越不稳定）：
// 无弱层时基础分 15；有弱层时 = 弱层分×0.55 + 板层势×(1+晶型加成)。
func ProfileStability(p *model.SnowProfile) (float64, WeakLayerCandidate, bool) {
	cand, found := FindWeakLayer(p.Layers)
	if !found {
		return 15, cand, false
	}
	slab := SlabScore(p.Layers, cand.Index)
	composite := cand.Score*0.55 + slab*(1+cand.Score/200)
	if composite > 100 {
		composite = 100
	}
	return composite, cand, true
}
