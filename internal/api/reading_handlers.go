package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// ingestReq 是批量上报请求体。
type ingestReq struct {
	StationID string        `json:"station_id"`
	Checksum  string        `json:"checksum"`
	Readings  []readingItem `json:"readings"`
}

// readingItem 是外部上报的读数条目。雪深字段允许站点以米为单位上报
// （unit=m），数据中心在边界处统一折算为厘米存储。
type readingItem struct {
	SensorKind string  `json:"sensor_kind"`
	Value      float64 `json:"value"`
	Unit       string  `json:"unit,omitempty"`
	Timestamp  string  `json:"timestamp"`
}

func (h *Handler) ingestBatch(w http.ResponseWriter, r *http.Request) {
	var req ingestReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, model.ErrEmptyBatch)
		return
	}
	batch := model.ReadingBatch{
		StationID: req.StationID,
		Checksum:  req.Checksum,
	}
	for _, item := range req.Readings {
		ts, err := time.Parse(time.RFC3339, item.Timestamp)
		if err != nil {
			writeErr(w, model.ErrReadingBadTime)
			return
		}
		value := item.Value
		if item.Unit == "m" && model.SensorKind(item.SensorKind) == model.SensorSnowDepth {
			value = value * 100 // 米 -> 厘米
		}
		batch.Readings = append(batch.Readings, model.Reading{
			StationID:  req.StationID,
			SensorKind: model.SensorKind(item.SensorKind),
			Value:      value,
			RecordedAt: ts.UTC(),
		})
	}
	alerts, err := h.svc.IngestBatch(r.Context(), &batch)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"accepted":     len(batch.Readings),
		"alerts_fired": alerts,
	})
}

func (h *Handler) latestReadings(w http.ResponseWriter, r *http.Request) {
	readings := h.svc.LatestReadings(r.PathValue("id"))
	if readings == nil {
		readings = []model.Reading{}
	}
	writeJSON(w, http.StatusOK, readings)
}

func (h *Handler) runEvaluation(w http.ResponseWriter, r *http.Request) {
	ev, err := h.svc.RunEvaluation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ev)
}

func (h *Handler) listEvaluations(w http.ResponseWriter, r *http.Request) {
	limit := queryIntDefault(r, "limit", 20)
	evs, err := h.svc.Store().LatestEvaluations(r.PathValue("id"), limit)
	if err != nil {
		writeErr(w, err)
		return
	}
	if evs == nil {
		evs = []model.StabilityEvaluation{}
	}
	writeJSON(w, http.StatusOK, evs)
}

func (h *Handler) dailySummary(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	from := now.AddDate(0, 0, -7)
	to := now
	var err error
	if fromStr != "" {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			writeErr(w, model.ErrReadingBadTime)
			return
		}
	}
	if toStr != "" {
		to, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			writeErr(w, model.ErrReadingBadTime)
			return
		}
	}
	rows, err := h.svc.StationDailySummary(r.PathValue("id"), from, to.Add(24*time.Hour))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func queryIntDefault(r *http.Request, key string, def int) int {
	v := strings.TrimSpace(r.URL.Query().Get(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}
