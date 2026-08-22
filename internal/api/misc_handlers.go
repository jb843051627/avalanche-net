package api

import (
	"net/http"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
	"github.com/jb843051627/avalanche-net/internal/service"
)

// queryOr 提取必填查询参数，缺失时直接回 400。
func queryOr(w http.ResponseWriter, r *http.Request, key string) string {
	v := r.URL.Query().Get(key)
	if v == "" {
		writeErr(w, model.ErrInvalidStation)
		return ""
	}
	return v
}

// parseWindowParams 解析 from/to（RFC3339），缺省最近 24 小时。
func parseWindowParams(fromStr, toStr string, now time.Time) (time.Time, time.Time, error) {
	return service.ParseExportWindow(fromStr, toStr, now)
}

func (h *Handler) overview(w http.ResponseWriter, r *http.Request) {
	ov, err := h.svc.Overview()
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ov)
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
