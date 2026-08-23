package regression

import (
	"testing"

	"github.com/jb843051627/avalanche-net/internal/engine"
	"github.com/jb843051627/avalanche-net/internal/model"
)

func TestBug27_PersistenceDecayKeepsDeepHoarDanger(t *testing.T) {
	p := &model.SnowProfile{ID: "p", Layers: []model.SnowLayer{
		{Index: 1, DepthFromCm: 0, DepthToCm: 45, DensityKgM3: 280, GrainShape: model.GrainFaceted, Hardness: model.HardnessFist, TempC: -7},
	}}
	res := engine.AssessPersistence(p, p.ObservedAt)
	if !res.Persistent || res.Index != 1 {
		t.Fatalf("fresh faceted slab weak layer not persistent: %+v", res)
	}
	if res.Score < 5 {
		t.Fatalf("persistence score %.2f collapsed below floor", res.Score)
	}
}
