package regression

import (
	"sync"
	"testing"

	"github.com/jb843051627/avalanche-net/internal/cache"
	"github.com/jb843051627/avalanche-net/internal/model"
)

func TestBug01_CacheConcurrentUpdateKeepsAllReadings(t *testing.T) {
	c := cache.New()
	c.Update("ST01", model.Reading{StationID: "ST01", SensorKind: model.SensorAirTemp, Value: -5}, 1000000)
	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				c.Update("ST01", model.Reading{StationID: "ST01", SensorKind: model.SensorSnowDepth, Value: float64(i)}, 1000000)
			}
		}()
	}
	wg.Wait()
	if got := len(c.Get("ST01")); got != 801 {
		t.Fatalf("cached readings = %d, want 801 (writes under read lock lose updates)", got)
	}
}
