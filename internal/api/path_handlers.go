package api

import (
	"net/http"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
	"github.com/jb843051627/avalanche-net/internal/service"
)

// PathExposureAlias 空列表 JSON 类型别名。
type PathExposureAlias = service.PathExposureRow

func (h *Handler) registerPath(w http.ResponseWriter, r *http.Request) {
	var p model.AvalanchePath
	if err := decodeJSON(r, &p); err != nil {
		writeErr(w, model.ErrPathInvalid)
		return
	}
	if err := h.svc.RegisterPath(&p); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *Handler) listPaths(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListPaths(queryOr(w, r, "region"))
	if err != nil {
		writeErr(w, err)
		return
	}
	if list == nil {
		list = []*model.AvalanchePath{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) pathExposure(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.PathExposureRanking(queryOr(w, r, "region"))
	if err != nil {
		writeErr(w, err)
		return
	}
	if rows == nil {
		rows = []PathExposureAlias{}
	}
	writeJSON(w, http.StatusOK, rows)
}

func (h *Handler) telemetryGaps(w http.ResponseWriter, r *http.Request) {
	stationID := r.PathValue("id")
	now := time.Now().UTC()
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	from := now.Add(-24 * time.Hour)
	var err error
	if fromStr != "" {
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			writeErr(w, model.ErrReadingBadTime)
			return
		}
	}
	to := now
	if toStr != "" {
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			writeErr(w, model.ErrReadingBadTime)
			return
		}
	}
	gaps, err := h.svc.DetectTelemetryGaps(stationID, from, to)
	if err != nil {
		writeErr(w, err)
		return
	}
	completeness, _ := h.svc.StationCompleteness(stationID, from, to)
	writeJSON(w, http.StatusOK, map[string]any{
		"gaps":         gaps,
		"completeness": completeness,
	})
}
