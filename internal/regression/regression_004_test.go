package regression

import (
	"fmt"
	"testing"
	"time"

	"github.com/jb843051627/avalanche-net/internal/cache"
	"github.com/jb843051627/avalanche-net/internal/model"
)

func TestBug04_CacheReadsSurviveConcurrentStationWrites(t *testing.T) {
	c := cache.New()
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			id := fmt.Sprintf("ST-%d", i)
			c.Update(id, model.Reading{StationID: id, SensorKind: model.SensorAirTemp, Value: -5}, 16)
		}
	}()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		_ = c.SortedStations()
		_ = c.Get("ST-0")
		_, _ = c.LatestByKind("ST-1", model.SensorAirTemp)
	}
	close(stop)
	<-done
}
