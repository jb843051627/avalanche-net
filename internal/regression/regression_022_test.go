package regression

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/jb843051627/avalanche-net/internal/model"
	"github.com/jb843051627/avalanche-net/internal/service"
	"github.com/jb843051627/avalanche-net/internal/store"
)

func TestBug22_LayerInsertFailureRollbacksProfile(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	svc := service.New(st)
	if err := svc.RegisterRegion("R1", "r"); err != nil {
		t.Fatal(err)
	}
	if err := svc.RegisterStation(&model.Station{ID: "S", Name: "n", RegionID: "R1",
		ElevationM: 2000, Aspect: model.AspectN, Lat: 1, Lon: 2}); err != nil {
		t.Fatal(err)
	}
	err = st.Transaction(func(tx *sql.Tx) error {
		_, err := tx.Exec(`INSERT INTO snow_layers(profile_id,idx,depth_from_cm,depth_to_cm,density_kgm3,grain_shape,hardness,temp_c)
			VALUES('P-C',1,0,10,200,'rounded','pencil',-3)`)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	layers := []model.SnowLayer{
		{Index: 1, DepthFromCm: 0, DepthToCm: 10, DensityKgM3: 220, GrainShape: model.GrainRounded, Hardness: model.Hardness1F, TempC: -2},
		{Index: 2, DepthFromCm: 10, DepthToCm: 40, DensityKgM3: 260, GrainShape: model.GrainFaceted, Hardness: model.HardnessFist, TempC: -4},
	}
	createErr := svc.CreateProfile(&model.SnowProfile{ID: "P-C", StationID: "S", Observer: "o", Layers: layers})
	if createErr == nil {
		t.Fatal("conflicting layer insert reported success")
	}
	if _, err := svc.GetProfile("P-C"); !errors.Is(err, model.ErrProfileNotFound) {
		t.Fatalf("partial profile visible after failed create: err=%v", err)
	}
}
