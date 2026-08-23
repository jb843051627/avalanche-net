package regression

import (
	"testing"

	"github.com/jb843051627/avalanche-net/internal/cache"
	"github.com/jb843051627/avalanche-net/internal/model"
)

func TestBug17_CacheGetReturnsDefensiveCopy(t *testing.T) {
	c := cache.New()
	c.Update("S", model.Reading{StationID: "S", SensorKind: model.SensorAirTemp, Value: -3}, 100)
	got := c.Get("S")
	got[0].Value = 777
	if again := c.Get("S"); again[0].Value == 777 {
		t.Fatal("external in-place mutation leaked into cache (no defensive copy)")
	}
}
