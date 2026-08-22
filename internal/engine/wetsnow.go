package engine

import (
	"github.com/jb843051627/avalanche-net/internal/model"
)

// WetSnowResult 是湿雪失稳判定的输出。
type WetSnowResult struct {
	Susceptible bool    `json:"susceptible"` // 是否进入湿雪失稳敏感区
	MoisturePct float64 `json:"moisture_pct"`
	AirTempC    float64 `json:"air_temp_c"`
	WetLayerIdx int     `json:"wet_layer_idx"` // 首个含水率越限的层序号（0=无）
	RainOnSnow  bool    `json:"rain_on_snow"`  // 雨夹雪事件
	Score       float64 `json:"score"`         // 湿雪附加分 0-30
}

// WetSnowThresholds 湿雪判定阈值。
type WetSnowThresholds struct {
	MoistureWarnPct float64 // 体积含水率警戒值
	TempWarmC       float64 // 气温温暖阈值
}

// DefaultWetSnowThresholds 默认阈值：含水率 12%、气温 0°C。
func DefaultWetSnowThresholds() WetSnowThresholds {
	return WetSnowThresholds{MoistureWarnPct: 12, TempWarmC: 0}
}

// AssessWetSnow 综合剖面含水率、气温与降水判定湿雪失稳风险：
// - 任一层含水率 ≥ 警戒值即标记 wet layer；
// - 气温 ≥ 0°C 且 24h 降水 ≥ 10mm 视为雨夹雪；
// - 附加分 = 越限层数 ×8 + 雨夹雪 10，封顶 30。
func AssessWetSnow(p *model.SnowProfile, w *model.WeatherSample) WetSnowResult {
	th := DefaultWetSnowThresholds()
	res := WetSnowResult{}
	wetLayers := 0
	for i, l := range p.Layers {
		if l.DensityKgM3 >= th.MoistureWarnPct*8 { // 密度近似代理：≥96 kg/m³ 且晶型为融冻
			if l.GrainShape == model.GrainMeltFreeze {
				if res.WetLayerIdx == 0 {
					res.WetLayerIdx = i + 1
				}
				wetLayers++
			}
		}
	}
	if w != nil {
		res.MoisturePct = w.Precip24hMm / 4 // 经验换算：24h 降水的 1/4 折算表层体积含水率增量
		res.AirTempC = w.AirTempC
		if w.AirTempC >= th.TempWarmC && w.Precip24hMm >= 10 {
			res.RainOnSnow = true
			wetLayers += 2
		}
	}
	res.Score = float64(wetLayers) * 8
	if res.Score > 30 {
		res.Score = 30
	}
	res.Susceptible = res.WetLayerIdx > 0 || res.RainOnSnow
	return res
}
