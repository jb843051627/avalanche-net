package engine

import (
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// PersistenceResult 是持续弱层指数（PWL）评估输出。
// 深霜/ facets 类晶型一旦被上覆载荷掩埋，可在数周内保持危险。
type PersistenceResult struct {
	Persistent    bool    `json:"persistent"`
	BurialAgeDays float64 `json:"burial_age_days"`
	Index         int     `json:"layer_idx"` // 弱层层序号
	Score         float64 `json:"score"`     // 持续度附加分 0-20
}

// GrainDecayDays 返回晶型的危险衰减周期（天）：超过该周期权重减半。
func GrainDecayDays(g model.GrainShape) float64 {
	switch g {
	case model.GrainDepthHoar:
		return 45
	case model.GrainFaceted:
		return 30
	case model.GrainMeltFreeze:
		return 10
	case model.GrainDecomposing:
		return 7
	default:
		return 3
	}
}

// AssessPersistence 以观测时间与弱层埋深估算弱层持续度：
// 掩埋越深、晶型越持久，附加分越高；衰减按半衰近似。
func AssessPersistence(p *model.SnowProfile, now time.Time) PersistenceResult {
	res := PersistenceResult{}
	cand, found := FindWeakLayer(p.Layers)
	if !found {
		return res
	}
	res.Index = cand.Index
	age := now.Sub(p.ObservedAt).Hours() / 24
	if age < 0 {
		age = 0
	}
	res.BurialAgeDays = age
	decay := GrainDecayDays(p.Layers[cand.Index-1].GrainShape)
	halfLives := age / decay
	decayFactor := 1.0
	for i := 0; i < int(halfLives) && i < 4; i++ {
		decayFactor /= 2
	}
	depthBoost := cand.DepthCm / 100 // 米化埋深加成
	if depthBoost > 1.5 {
		depthBoost = 1.5
	}
	res.Score = round2f(cand.Score * 0.2 * decayFactor * (0.5 + depthBoost))
	if res.Score > 20 {
		res.Score = 20
	}
	res.Persistent = res.Score >= 5
	return res
}

func round2f(v float64) float64 { return float64(int(v*100+0.5)) / 100 }
