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

func TestBug09_SentinelErrorsSurviveWrapping(t *testing.T) {
	svc := newSvc(t)
	seedRegionStation(t, svc, "ST09")
	dup := &model.Station{ID: "other", Name: "x", RegionID: "R1", ElevationM: 2000, Aspect: model.AspectE, Lat: 1, Lon: 2}
	if err := svc.RegisterStation(dup); err != nil {
		t.Fatal(err)
	}
	if err := svc.RegisterStation(dup); !errors.Is(err, model.ErrStationExists) {
		t.Fatalf("duplicate register err=%v, want ErrStationExists chain", err)
	}
	if err := svc.SetStationStatus("missing-st", model.StatusOnline); !errors.Is(err, model.ErrStationNotFound) {
		t.Fatalf("missing station status err=%v, want ErrStationNotFound chain", err)
	}
}
