package regression

import (
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

func TestBug10_FailedPersistenceNotSwallowed(t *testing.T) {
	svc := newSvc(t)
	seedRegionStation(t, svc, "ST10")
	layers := []model.SnowLayer{
		{Index: 1, DepthFromCm: 0, DepthToCm: 30, DensityKgM3: 250, GrainShape: model.GrainRounded, Hardness: model.Hardness1F, TempC: -2},
	}
	first := &model.SnowProfile{ID: "P-DUP", StationID: "ST10", Observer: "o", Layers: layers}
	if err := svc.CreateProfile(first); err != nil {
		t.Fatal(err)
	}
	err := svc.CreateProfile(&model.SnowProfile{ID: "P-DUP", StationID: "ST10", Observer: "o", Layers: layers})
	if err == nil {
		t.Fatal("duplicate profile insert reported success (store error swallowed)")
	}
}
