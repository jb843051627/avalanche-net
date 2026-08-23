package regression

import (
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

func TestBug13_MissingStationRegistrationStillAllowed(t *testing.T) {
	svc := newSvc(t)
	seedRegionStation(t, svc, "ST13")
	next := &model.Station{ID: "ST13B", Name: "hut-b", RegionID: "R1", ElevationM: 2100, Aspect: model.AspectW, Lat: 43.2, Lon: 86.9}
	if err := svc.RegisterStation(next); err != nil {
		t.Fatalf("registering a brand-new station failed: %v", err)
	}
	err := svc.SetStationStatus("no-such-station", model.StatusOnline)
	if !errors.Is(err, model.ErrStationNotFound) {
		t.Fatalf("missing station status err=%v, want ErrStationNotFound", err)
	}
}
