package regression

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/jb843051627/avalanche-net/internal/engine"
	"github.com/jb843051627/avalanche-net/internal/model"
	"github.com/jb843051627/avalanche-net/internal/service"
	"github.com/jb843051627/avalanche-net/internal/store"
)

func newSvc23(t *testing.T) *service.Service {
	st, err := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return service.New(st)
}

func seedRS23(t *testing.T, svc *service.Service) {
	if err := svc.RegisterRegion("R1", "r"); err != nil {
		t.Fatal(err)
	}
	if err := svc.RegisterStation(&model.Station{ID: "S", Name: "n", RegionID: "R1",
		ElevationM: 2000, Aspect: model.AspectN, Lat: 1, Lon: 2}); err != nil {
		t.Fatal(err)
	}
}

func TestBug23_DedupWindowAndReminderIntervalsPositive(t *testing.T) {
	svc := newSvc23(t)
	seedRS23(t, svc)
	cand := engine.AlertCandidate{RuleKey: "windspeed:crit", Level: model.LevelCritical, Reason: "gust", Value: 95}
	now := time.Now().UTC()
	first, err := svc.RaiseAlert(context.Background(), "S", cand, now)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.RaiseAlert(context.Background(), "S", cand, now.Add(time.Minute))
	if !errors.Is(err, model.ErrDeduplicated) {
		t.Fatalf("second alert within window err=%v first=%v, want ErrDeduplicated", err, first.ID)
	}
	fresh := &model.Alert{StationID: "S", RuleKey: "k", Level: model.LevelCritical, State: model.StateActive,
		TriggeredAt: now}
	if engine.NeedsEscalationReminder(fresh, now.Add(time.Minute)) {
		t.Fatal("fresh critical alert must not be escalated before reminder interval")
	}
}
