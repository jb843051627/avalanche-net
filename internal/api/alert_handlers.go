package api

import (
	"net/http"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

func (h *Handler) listAlerts(w http.ResponseWriter, r *http.Request) {
	alerts, err := h.svc.ListActiveAlerts(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	if alerts == nil {
		alerts = []*model.Alert{}
	}
	writeJSON(w, http.StatusOK, alerts)
}

type ackReq struct {
	AckedBy string `json:"acked_by"`
}

func (h *Handler) ackAlert(w http.ResponseWriter, r *http.Request) {
	var req ackReq
	_ = decodeJSON(r, &req)
	a, err := h.svc.AckAlert(r.PathValue("id"), req.AckedBy)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *Handler) resolveAlert(w http.ResponseWriter, r *http.Request) {
	a, err := h.svc.ResolveAlert(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *Handler) draftBulletin(w http.ResponseWriter, r *http.Request) {
	var b model.Bulletin
	if err := decodeJSON(r, &b); err != nil {
		writeErr(w, model.ErrInvalidDanger)
		return
	}
	b.Stage = model.BulletinDraft
	if err := h.svc.DraftBulletin(&b); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

func (h *Handler) listBulletins(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListPublishedBulletins(r.URL.Query().Get("region"))
	if err != nil {
		writeErr(w, err)
		return
	}
	if list == nil {
		list = []*model.Bulletin{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) publishBulletin(w http.ResponseWriter, r *http.Request) {
	b, err := h.svc.PublishBulletin(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) archiveBulletin(w http.ResponseWriter, r *http.Request) {
	b, err := h.svc.ArchiveBulletin(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) exportReadings(w http.ResponseWriter, r *http.Request) {
	stationID := queryOr(w, r, "station")
	if stationID == "" {
		return
	}
	from, to, ok := parseExportWindowOrBadRequest(w, r)
	if !ok {
		return
	}
	csvBytes, err := h.svc.ExportReadingsCSV(stationID, from, to)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeCSV(w, csvBytes)
}

func (h *Handler) exportAlertsCSV(w http.ResponseWriter, r *http.Request) {
	stationID := queryOr(w, r, "station")
	if stationID == "" {
		return
	}
	csvBytes, err := h.svc.ExportAlertsCSV(stationID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeCSV(w, csvBytes)
}

func writeCSV(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func parseExportWindowOrBadRequest(w http.ResponseWriter, r *http.Request) (time.Time, time.Time, bool) {
	now := time.Now().UTC()
	from, to, err := parseWindowParams(r.URL.Query().Get("from"), r.URL.Query().Get("to"), now)
	if err != nil {
		writeErr(w, model.ErrReadingBadTime)
		return time.Time{}, time.Time{}, false
	}
	return from, to, true
}
