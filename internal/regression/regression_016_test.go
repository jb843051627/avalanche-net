package regression

import (
	"testing"

	"github.com/jb843051627/avalanche-net/internal/engine"
	"github.com/jb843051627/avalanche-net/internal/model"
)

func TestBug16_CombineDangerNilWeatherSafe(t *testing.T) {
	score, level := engine.CombineDanger(40, nil, model.AspectN)
	if level != model.DangerModerate || score <= 0 {
		t.Fatalf("nil weather combine got (%.1f,%s), want moderate without panic", score, level)
	}
}
