package regression

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/jb843051627/avalanche-net/internal/service"
	"github.com/jb843051627/avalanche-net/internal/store"
)

func TestBug30_LoadingSummaryHandlesEmptyRegion(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	svc := service.New(st)
	sum, err := svc.RegionLoadingSummary("empty-region", 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if sum.Samples != 0 || sum.MaxWindKmh != 0 {
		t.Fatalf("empty region summary polluted: %+v", sum)
	}
}
