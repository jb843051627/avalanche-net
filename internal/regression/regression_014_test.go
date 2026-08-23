package regression

import (
	"testing"

	"github.com/jb843051627/avalanche-net/internal/engine"
	"github.com/jb843051627/avalanche-net/internal/model"
)

func TestBug14_EmptyProfileStabilityNoWeakClaim(t *testing.T) {
	score, cand, hasWeak := engine.ProfileStability(&model.SnowProfile{ID: "empty"})
	if hasWeak || cand.Index != 0 {
		t.Fatalf("empty profile fabricated weak layer idx=%d score=%.1f", cand.Index, score)
	}
	dry := []model.SnowLayer{{Index: 1, DepthFromCm: 0, DepthToCm: 80, DensityKgM3: 320,
		GrainShape: model.GrainRounded, Hardness: model.HardnessPenc, TempC: -8}}
	_, cand2, hasWeak2 := engine.ProfileStability(&model.SnowProfile{ID: "dry", Layers: dry})
	if hasWeak2 || cand2.Index != 0 {
		t.Fatalf("benign rounded layer flagged as weak layer idx=%d", cand2.Index)
	}
}
