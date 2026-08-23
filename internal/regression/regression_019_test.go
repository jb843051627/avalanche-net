package regression

import (
	"path/filepath"
	"testing"

	"github.com/jb843051627/avalanche-net/internal/model"
	"github.com/jb843051627/avalanche-net/internal/service"
	"github.com/jb843051627/avalanche-net/internal/store"
)

func TestBug19_StatusMachinesRejectIllegalJumps(t *testing.T) {
	st := &model.Station{ID: "x", Status: model.StatusOffline}
	if st.CanTransition(model.StatusMaintenance) {
		t.Fatal("offline -> maintenance must be rejected")
	}
	b := &model.Bulletin{ID: "b", Stage: model.BulletinArchived}
	if b.Stage.CanTransition(model.BulletinPublished) {
		t.Fatal("archived -> published must be rejected")
	}
	dbst, err := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer dbst.Close()
	svc := service.New(dbst)
	if err := svc.RegisterRegion("R1", "r"); err != nil {
		t.Fatal(err)
	}
	if err := svc.RegisterStation(&model.Station{ID: "S", Name: "n", RegionID: "R1",
		ElevationM: 2000, Aspect: model.AspectN, Lat: 1, Lon: 2}); err != nil {
		t.Fatal(err)
	}
	if err := svc.SetStationStatus("S", model.StatusMaintenance); err == nil {
		t.Fatal("offline -> maintenance accepted via service")
	}
}
