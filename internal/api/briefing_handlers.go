package api

import (
	"net/http"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
	"github.com/jb843051627/avalanche-net/internal/service"
)

// WeeklyTrendAlias 空列表 JSON 类型别名。
type WeeklyTrendAlias = service.WeeklyTrendRow

func (h *Handler) bulletinText(w http.ResponseWriter, r *http.Request) {
	regionID := queryOr(w, r, "region")
	if regionID == "" {
		return
	}
	issuedFor := time.Now().UTC()
	if v := r.URL.Query().Get("date"); v != "" {
		parsed, err := time.Parse("2006-01-02", v)
		if err != nil {
			writeErr(w, model.ErrReadingBadTime)
			return
		}
		issuedFor = parsed
	}
	text, err := h.svc.BuildBulletinText(regionID, issuedFor)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, text)
}

func (h *Handler) autoDraftBulletin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RegionID  string `json:"region_id"`
		IssuedFor string `json:"issued_for"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, model.ErrInvalidStation)
		return
	}
	issuedFor := time.Now().UTC()
	if req.IssuedFor != "" {
		parsed, err := time.Parse("2006-01-02", req.IssuedFor)
		if err != nil {
			writeErr(w, model.ErrReadingBadTime)
			return
		}
		issuedFor = parsed
	}
	b, err := h.svc.AutoDraftBulletin(req.RegionID, issuedFor)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

func (h *Handler) weeklyTrend(w http.ResponseWriter, r *http.Request) {
	days := queryIntDefault(r, "days", 7)
	rows, err := h.svc.StationWeeklyTrend(r.PathValue("id"), days)
	if err != nil {
		writeErr(w, err)
		return
	}
	if rows == nil {
		rows = []WeeklyTrendAlias{}
	}
	writeJSON(w, http.StatusOK, rows)
}

func (h *Handler) exportSummaryCSV(w http.ResponseWriter, r *http.Request) {
	stationID := r.PathValue("id")
	now := time.Now().UTC()
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	from := now.AddDate(0, 0, -7)
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
	rows, err := h.svc.StationDailySummary(stationID, from, to)
	if err != nil {
		writeErr(w, err)
		return
	}
	csvBytes := h.svc.DailySummaryCSV(rows)
	writeCSV(w, csvBytes)
}
