package regression

import (
	"reflect"
	"testing"

	"github.com/jb843051627/avalanche-net/internal/model"
	"github.com/jb843051627/avalanche-net/internal/service"
)

func TestBug18_NormalizeLayersKeepsCallerSliceIntact(t *testing.T) {
	in := []model.SnowLayer{
		{Index: 1, DepthFromCm: 10, DepthToCm: 30},
		{Index: 2, DepthFromCm: 0, DepthToCm: 10},
	}
	want := append([]model.SnowLayer(nil), in...)
	out := service.NormalizeLayers(in)
	if out[0].DepthFromCm != 0 {
		t.Fatalf("normalize did not sort ascending: %+v", out)
	}
	if !reflect.DeepEqual(in, want) {
		t.Fatalf("caller slice mutated in place: %+v", in)
	}
}
