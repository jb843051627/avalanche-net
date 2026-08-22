package api

import (
	"net/http"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

func (h *Handler) createProfile(w http.ResponseWriter, r *http.Request) {
	var p model.SnowProfile
	if err := decodeJSON(r, &p); err != nil {
		writeErr(w, model.ErrNoLayers)
		return
	}
	p.Layers = normalizeProfileLayers(p.Layers)
	if err := h.svc.CreateProfile(&p); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

// normalizeProfileLayers 保证 index 连续且按深度升序。
func normalizeProfileLayers(layers []model.SnowLayer) []model.SnowLayer {
	if len(layers) < 2 {
		return layers
	}
	for i := 1; i < len(layers); i++ {
		for j := i; j > 0 && layers[j].DepthFromCm < layers[j-1].DepthFromCm; j-- {
			layers[j], layers[j-1] = layers[j-1], layers[j]
		}
	}
	for i := range layers {
		layers[i].Index = i + 1
	}
	return layers
}

func (h *Handler) getProfile(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.GetProfile(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) latestProfile(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.LatestProfile(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) scheduleInspection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StationID string `json:"station_id"`
		DueDate   string `json:"due_date"`
		Assignee  string `json:"assignee"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, model.ErrInvalidStation)
		return
	}
	due, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil {
		writeErr(w, model.ErrReadingBadTime)
		return
	}
	task, err := h.svc.ScheduleInspection(req.StationID, due, req.Assignee)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (h *Handler) listInspections(w http.ResponseWriter, r *http.Request) {
	status := model.InspectionStatus(r.URL.Query().Get("status"))
	switch status {
	case "", model.InspectionDue, model.InspectionInProgress,
		model.InspectionCompleted, model.InspectionCancelled:
	default:
		writeErr(w, model.ErrInspectionState)
		return
	}
	tasks, err := h.svc.ListInspections(r.URL.Query().Get("station"), status)
	if err != nil {
		writeErr(w, err)
		return
	}
	if tasks == nil {
		tasks = []*model.InspectionTask{}
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *Handler) startInspection(w http.ResponseWriter, r *http.Request) {
	t, err := h.svc.StartInspection(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

type completeReq struct {
	Notes string `json:"notes"`
}

func (h *Handler) completeInspection(w http.ResponseWriter, r *http.Request) {
	var req completeReq
	if err := decodeJSON(r, &req); err != nil && err.Error() != "EOF" {
		writeErr(w, model.ErrInvalidStation)
		return
	}
	t, err := h.svc.CompleteInspection(r.PathValue("id"), req.Notes)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}
