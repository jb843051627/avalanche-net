package regression

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/jb843051627/avalanche-net/internal/model"
	"github.com/jb843051627/avalanche-net/internal/service"
	"github.com/jb843051627/avalanche-net/internal/store"
)

func newSvc(t *testing.T) *service.Service {
	st, err := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return service.New(st)
}

func seedRegionStation(t *testing.T, svc *service.Service, id string) {
	if err := svc.RegisterRegion("R1", "north slope"); err != nil {
		t.Fatal(err)
	}
	st := &model.Station{ID: id, Name: "hut-" + id, RegionID: "R1", ElevationM: 2400, Aspect: model.AspectN, Lat: 43.1, Lon: 86.8}
	if err := svc.RegisterStation(st); err != nil {
		t.Fatal(err)
	}
}

func TestBug06_RunEvaluationCanceledContextPropagates(t *testing.T) {
	svc := newSvc(t)
	seedRegionStation(t, svc, "ST06")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := svc.RunEvaluation(ctx, "ST06")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v, want context.Canceled (evaluation detaches caller cancellation)", err)
	}
}
