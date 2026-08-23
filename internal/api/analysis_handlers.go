package api

import (
	"net/http"
	"time"

	"github.com/jb843051627/avalanche-net/internal/engine"
	"github.com/jb843051627/avalanche-net/internal/model"
	"github.com/jb843051627/avalanche-net/internal/service"
)

// DriftReportAlias / RoseCellAlias 仅为 JSON 空列表类型别名。
type DriftReportAlias = service.DriftReport
type RoseCellAlias = engine.RoseCell

func (h *Handler) recordWeather(w http.ResponseWriter, r *http.Request) {
	var wreq model.WeatherSample
	if err := decodeJSON(r, &wreq); err != nil {
		writeErr(w, model.ErrInvalidStation)
		return
	}
	if err := h.svc.RecordWeatherSample(&wreq); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, wreq)
}

func (h *Handler) latestWeather(w http.ResponseWriter, r *http.Request) {
	sample, err := h.svc.LatestWeather(queryOr(w, r, "region"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sample)
}

func (h *Handler) loadingSummary(w http.ResponseWriter, r *http.Request) {
	regionID := queryOr(w, r, "region")
	if regionID == "" {
		return
	}
	sum, err := h.svc.RegionLoadingSummary(regionID, 24*time.Hour)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sum)
}

type calibrationReq struct {
	StationID    string  `json:"station_id"`
	SensorKind   string  `json:"sensor_kind"`
	ReferenceIn  float64 `json:"reference_in"`
	ReportedOut  float64 `json:"reported_out"`
	CalibratedBy string  `json:"calibrated_by"`
}

func (h *Handler) recordCalibration(w http.ResponseWriter, r *http.Request) {
	var req calibrationReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, model.ErrInvalidStation)
		return
	}
	c, err := h.svc.RecordCalibration(req.StationID, model.SensorKind(req.SensorKind),
		req.ReferenceIn, req.ReportedOut, req.CalibratedBy)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *Handler) driftReport(w http.ResponseWriter, r *http.Request) {
	report, err := h.svc.StationDriftReport(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	if report == nil {
		report = []DriftReportAlias{}
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *Handler) recordObservation(w http.ResponseWriter, r *http.Request) {
	var o model.Observation
	if err := decodeJSON(r, &o); err != nil {
		writeErr(w, model.ErrObservationInvalid)
		return
	}
	if err := h.svc.RecordObservation(&o); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, o)
}

func (h *Handler) listObservations(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListObservations(queryOr(w, r, "region"))
	if err != nil {
		writeErr(w, err)
		return
	}
	if list == nil {
		list = []*model.Observation{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) regionRose(w http.ResponseWriter, r *http.Request) {
	cells, err := h.svc.BuildRegionRose(queryOr(w, r, "region"))
	if err != nil {
		writeErr(w, err)
		return
	}
	if cells == nil {
		cells = []RoseCellAlias{}
	}
	writeJSON(w, http.StatusOK, cells)
}

func (h *Handler) compareProfiles(w http.ResponseWriter, r *http.Request) {
	fromID := r.URL.Query().Get("from")
	toID := r.URL.Query().Get("to")
	if fromID == "" || toID == "" {
		writeErr(w, model.ErrProfileNotFound)
		return
	}
	diff, err := h.svc.CompareStationProfiles(fromID, toID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, diff)
}
