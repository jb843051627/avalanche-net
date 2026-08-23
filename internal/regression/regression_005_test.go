package regression

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
	"github.com/jb843051627/avalanche-net/internal/service"
	"github.com/jb843051627/avalanche-net/internal/store"
	"github.com/jb843051627/avalanche-net/internal/validation"
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

func TestBug05_IngestBatchCanceledContextAborts(t *testing.T) {
	svc := newSvc(t)
	seedRegionStation(t, svc, "ST05")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	batch := &model.ReadingBatch{StationID: "ST05", Readings: []model.Reading{
		{StationID: "ST05", SensorKind: model.SensorSnowDepth, Value: 42, RecordedAt: time.Now()},
	}}
	batch.Checksum = validation.Sign(service.BatchFingerprint(batch))
	n, err := svc.IngestBatch(ctx, batch)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v accepted=%d, want context.Canceled", err, n)
	}
}
