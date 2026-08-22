package api

import (
	"net/http"

	"github.com/jb843051627/avalanche-net/internal/model"
)

type regionReq struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h *Handler) createRegion(w http.ResponseWriter, r *http.Request) {
	var req regionReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, model.ErrInvalidStation)
		return
	}
	if req.ID == "" || req.Name == "" {
		writeErr(w, model.ErrInvalidStation)
		return
	}
	if err := h.svc.RegisterRegion(req.ID, req.Name); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": req.ID})
}

func (h *Handler) listRegions(w http.ResponseWriter, r *http.Request) {
	regions, err := h.svc.ListRegions()
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, regions)
}

func (h *Handler) createStation(w http.ResponseWriter, r *http.Request) {
	var st model.Station
	if err := decodeJSON(r, &st); err != nil {
		writeErr(w, model.ErrInvalidStation)
		return
	}
	if err := h.svc.RegisterStation(&st); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, st)
}

func (h *Handler) listStations(w http.ResponseWriter, r *http.Request) {
	stations, err := h.svc.ListStations(r.URL.Query().Get("region"))
	if err != nil {
		writeErr(w, err)
		return
	}
	if stations == nil {
		stations = []*model.Station{}
	}
	writeJSON(w, http.StatusOK, stations)
}

func (h *Handler) getStation(w http.ResponseWriter, r *http.Request) {
	st, err := h.svc.GetStation(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

type statusReq struct {
	Status string `json:"status"`
}

func (h *Handler) setStationStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req statusReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, model.ErrInvalidStation)
		return
	}
	to := model.StationStatus(req.Status)
	switch to {
	case model.StatusOnline, model.StatusOffline, model.StatusMaintenance:
	default:
		writeErr(w, model.ErrInvalidStatusMove)
		return
	}
	if err := h.svc.SetStationStatus(id, to); err != nil {
		writeErr(w, err)
		return
	}
	st, _ := h.svc.GetStation(id)
	writeJSON(w, http.StatusOK, st)
}

func (h *Handler) stationHeartbeat(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.Heartbeat(id); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listSensors(w http.ResponseWriter, r *http.Request) {
	sensors, err := h.svc.ListSensors(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sensors)
}

func (h *Handler) configureSensor(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	kind := model.SensorKind(r.PathValue("kind"))
	var req struct {
		WarnAt float64 `json:"warn_at"`
		CritAt float64 `json:"crit_at"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, model.ErrInvalidStation)
		return
	}
	if err := h.svc.ConfigureSensor(id, kind, req.WarnAt, req.CritAt); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
