package regression

import (
	"testing"
	"time"

	"github.com/jb843051627/avalanche-net/internal/service"
)

func TestBug11_ParseExportWindowValidatesAndNormalizes(t *testing.T) {
	if _, _, err := service.ParseExportWindow("not-a-time", "", time.Now()); err == nil {
		t.Fatal("bad timestamp accepted silently")
	}
	now := time.Now().UTC()
	later := now.Add(2 * time.Hour).Format(time.RFC3339)
	earlier := now.Format(time.RFC3339)
	from, to, err := service.ParseExportWindow(later, earlier, now)
	if err != nil {
		t.Fatal(err)
	}
	if !from.Before(to) {
		t.Fatalf("window not normalized ascending: %v >= %v", from, to)
	}
}
