package regression

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jb843051627/avalanche-net/internal/service"
	"github.com/jb843051627/avalanche-net/internal/store"
)

func newSvc7(t *testing.T) *service.Service {
	st, err := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return service.New(st)
}

func TestBug07_HeartbeatStopIdempotent(t *testing.T) {
	svc := newSvc7(t)
	stop := svc.StartHeartbeatWatcher(context.Background())
	stop()
	stop()
}
