package regression

import (
	"strings"
	"testing"
	"time"

	"github.com/jb843051627/avalanche-net/internal/engine"
	"github.com/jb843051627/avalanche-net/internal/model"
)

func TestBug28_BulletinTextUsesPerBandLevels(t *testing.T) {
	entries := []engine.RoseEntry{
		{Aspect: model.AspectS, ElevationM: 2500, Level: model.DangerHigh},
		{Aspect: model.AspectN, ElevationM: 2400, Level: model.DangerConsiderable},
		{Aspect: model.AspectW, ElevationM: 1500, Level: model.DangerLow},
	}
	text := engine.GenerateBulletinText("R1", time.Now().UTC(), engine.BuildRose(entries), nil)
	if len(text.Sections) < 3 {
		t.Fatalf("expected per-band sections, got %d", len(text.Sections))
	}
	if !strings.Contains(text.Sections[0], "high") {
		t.Fatalf("above-treeline section lost its own band level: %q", text.Sections[0])
	}
	joined := strings.Join(text.Sections, "\n")
	if !strings.Contains(joined, "深霜弱层") {
		t.Fatalf("north-aspect considerable signal dropped from bulletin: %q", joined)
	}
}
