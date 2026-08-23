package api

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/jb843051627/avalanche-net/internal/service"
)

// Handler 持有业务服务，实现全部 HTTP 路由。
type Handler struct {
	svc *service.Service
}

// New 构造 HTTP handler。
func New(svc *service.Service) http.Handler {
	h := &Handler{svc: svc}
	mux := http.NewServeMux()

	// 站点
	mux.HandleFunc("GET /api/v1/regions", h.listRegions)
	mux.HandleFunc("POST /api/v1/regions", h.createRegion)
	mux.HandleFunc("POST /api/v1/stations", h.createStation)
	mux.HandleFunc("GET /api/v1/stations", h.listStations)
	mux.HandleFunc("GET /api/v1/stations/{id}", h.getStation)
	mux.HandleFunc("POST /api/v1/stations/{id}/status", h.setStationStatus)
	mux.HandleFunc("POST /api/v1/stations/{id}/heartbeat", h.stationHeartbeat)
	mux.HandleFunc("GET /api/v1/stations/{id}/sensors", h.listSensors)
	mux.HandleFunc("PUT /api/v1/stations/{id}/sensors/{kind}", h.configureSensor)

	// 采集
	mux.HandleFunc("POST /api/v1/ingest/batch", h.ingestBatch)
	mux.HandleFunc("GET /api/v1/stations/{id}/readings/latest", h.latestReadings)

	// 剖面
	mux.HandleFunc("POST /api/v1/profiles", h.createProfile)
	mux.HandleFunc("GET /api/v1/profiles/{id}", h.getProfile)
	mux.HandleFunc("GET /api/v1/stations/{id}/profiles/latest", h.latestProfile)

	// 评估
	mux.HandleFunc("POST /api/v1/stations/{id}/evaluate", h.runEvaluation)
	mux.HandleFunc("GET /api/v1/stations/{id}/evaluations", h.listEvaluations)

	// 告警
	mux.HandleFunc("GET /api/v1/stations/{id}/alerts", h.listAlerts)
	mux.HandleFunc("POST /api/v1/alerts/{id}/ack", h.ackAlert)
	mux.HandleFunc("POST /api/v1/alerts/{id}/resolve", h.resolveAlert)

	// 巡检
	mux.HandleFunc("POST /api/v1/inspections", h.scheduleInspection)
	mux.HandleFunc("GET /api/v1/inspections", h.listInspections)
	mux.HandleFunc("POST /api/v1/inspections/{id}/start", h.startInspection)
	mux.HandleFunc("POST /api/v1/inspections/{id}/complete", h.completeInspection)

	// 公报
	mux.HandleFunc("POST /api/v1/bulletins", h.draftBulletin)
	mux.HandleFunc("GET /api/v1/bulletins", h.listBulletins)
	mux.HandleFunc("POST /api/v1/bulletins/{id}/publish", h.publishBulletin)
	mux.HandleFunc("POST /api/v1/bulletins/{id}/archive", h.archiveBulletin)

	// 导出与统计
	mux.HandleFunc("GET /api/v1/export/readings.csv", h.exportReadings)
	mux.HandleFunc("GET /api/v1/export/alerts.csv", h.exportAlertsCSV)
	mux.HandleFunc("GET /api/v1/stations/{id}/summary/daily", h.dailySummary)
	mux.HandleFunc("GET /api/v1/overview", h.overview)
	mux.HandleFunc("GET /healthz", h.healthz)

	// 气象/校准/目击/玫瑰图
	mux.HandleFunc("POST /api/v1/weather", h.recordWeather)
	mux.HandleFunc("GET /api/v1/weather/latest", h.latestWeather)
	mux.HandleFunc("GET /api/v1/weather/loading", h.loadingSummary)
	mux.HandleFunc("POST /api/v1/calibrations", h.recordCalibration)
	mux.HandleFunc("GET /api/v1/stations/{id}/drift", h.driftReport)
	mux.HandleFunc("POST /api/v1/observations", h.recordObservation)
	mux.HandleFunc("GET /api/v1/observations", h.listObservations)
	mux.HandleFunc("GET /api/v1/rose", h.regionRose)
	mux.HandleFunc("GET /api/v1/profiles/compare", h.compareProfiles)

	// 路径与断档
	mux.HandleFunc("POST /api/v1/paths", h.registerPath)
	mux.HandleFunc("GET /api/v1/paths", h.listPaths)
	mux.HandleFunc("GET /api/v1/paths/exposure", h.pathExposure)
	mux.HandleFunc("GET /api/v1/stations/{id}/gaps", h.telemetryGaps)
	mux.HandleFunc("GET /api/v1/bulletins/text", h.bulletinText)
	mux.HandleFunc("POST /api/v1/bulletins/auto-draft", h.autoDraftBulletin)
	mux.HandleFunc("GET /api/v1/stations/{id}/trend/weekly", h.weeklyTrend)
	mux.HandleFunc("GET /api/v1/stations/{id}/summary/daily.csv", h.exportSummaryCSV)

	staticDir := locateStaticDir()
	if staticDir != "" {
		fs := http.FileServer(http.Dir(staticDir))
		mux.Handle("GET /", fs)
	}
	return Chain(mux, LogMiddleware, RecoverMiddleware)
}

// locateStaticDir 定位内嵌控制台静态目录（工作目录优先，其次 web/static）。
func locateStaticDir() string {
	for _, candidate := range []string{
		filepath.Join("web", "static"),
		filepath.Join("..", "web", "static"),
	} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return ""
}
