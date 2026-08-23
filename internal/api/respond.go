package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// writeJSON 输出 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

type errorBody struct {
	Error string `json:"error"`
}

// writeErr 把业务错误映射为 HTTP 状态码。
func writeErr(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, model.ErrStationNotFound),
		errors.Is(err, model.ErrProfileNotFound),
		errors.Is(err, model.ErrAlertNotFound),
		errors.Is(err, model.ErrInspectionNotFound),
		errors.Is(err, model.ErrBulletinNotFound):
		status = http.StatusNotFound
	case errors.Is(err, model.ErrInvalidStation),
		errors.Is(err, model.ErrReadingOutOfRange),
		errors.Is(err, model.ErrReadingBadTime),
		errors.Is(err, model.ErrReadingUnknownKind),
		errors.Is(err, model.ErrEmptyBatch),
		errors.Is(err, model.ErrNoLayers),
		errors.Is(err, model.ErrLayerOrder),
		errors.Is(err, model.ErrInvalidDanger),
		errors.Is(err, model.ErrNotesRequired),
		errors.Is(err, model.ErrInvalidStatusMove),
		errors.Is(err, model.ErrAlertState),
		errors.Is(err, model.ErrInspectionState),
		errors.Is(err, model.ErrBulletinStage):
		status = http.StatusBadRequest
	case errors.Is(err, model.ErrStationExists):
		status = http.StatusConflict
	case errors.Is(err, model.ErrDeduplicated):
		status = http.StatusOK // 去重命中不算错误，返回既有告警
	}
	writeJSON(w, status, errorBody{Error: err.Error()})
}

// decodeJSON 解析请求体到 out。
func decodeJSON(r *http.Request, out any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	return dec.Decode(out)
}
