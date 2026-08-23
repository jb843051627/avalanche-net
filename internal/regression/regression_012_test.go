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

func TestBug12_ChecksumAndWindowRejectionsEnforced(t *testing.T) {
	svc := newSvc(t)
	seedRegionStation(t, svc, "ST12")
	bad := &model.ReadingBatch{StationID: "ST12", Checksum: "deadbeef", Readings: []model.Reading{
		{StationID: "ST12", SensorKind: model.SensorSnowDepth, Value: 40, RecordedAt: time.Now()},
	}}
	if _, err := svc.IngestBatch(context.Background(), bad); !errors.Is(err, validation.ErrChecksumMismatch) {
		t.Fatalf("bad checksum err=%v, want ErrChecksumMismatch", err)
	}
	good := &model.ReadingBatch{StationID: "ST12", Readings: []model.Reading{
		{StationID: "ST12", SensorKind: model.SensorSnowDepth, Value: 41, RecordedAt: time.Now().Add(-3 * time.Hour)},
	}}
	good.Checksum = validation.Sign(service.BatchFingerprint(good))
	_, err := svc.IngestBatch(context.Background(), good)
	if err == nil {
		t.Fatal("stale-window batch accepted")
	}
	if n := len(svc.LatestReadings("ST12")); n != 0 {
		t.Fatalf("stale reading stored, cache=%d", n)
	}
}
