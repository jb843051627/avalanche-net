package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// ExportReadingsCSV 导出站点窗口内读数为 CSV（时间一律 UTC RFC3339）。
func (s *Service) ExportReadingsCSV(stationID string, from, to time.Time) ([]byte, error) {
	if _, err := s.store.GetStation(stationID); err != nil {
		return nil, err
	}
	rows, err := s.store.ListReadings(stationID, from.UTC(), to.UTC())
	if err != nil {
		return nil, err
	}
	var sb strings.Builder
	sb.WriteString("station_id,sensor_kind,value,recorded_at_utc\n")
	for _, r := range rows {
		sb.WriteString(fmt.Sprintf("%s,%s,%s,%s\n",
			r.StationID, string(r.SensorKind), formatFloat(r.Value),
			r.RecordedAt.UTC().Format(time.RFC3339)))
	}
	return []byte(sb.String()), nil
}

// ExportAlertsCSV 导出站点告警历史为 CSV。
func (s *Service) ExportAlertsCSV(stationID string) ([]byte, error) {
	alerts, err := s.store.ListAlertsByStation(stationID, "")
	if err != nil {
		return nil, err
	}
	var sb strings.Builder
	sb.WriteString("id,station_id,rule_key,level,state,value,triggered_at_utc\n")
	for _, a := range alerts {
		sb.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s\n",
			a.ID, a.StationID, a.RuleKey, string(a.Level), string(a.State),
			formatFloat(a.Value), a.TriggeredAt.UTC().Format(time.RFC3339)))
	}
	return []byte(sb.String()), nil
}

// ExportEvaluationsCSV 导出站点评估趋势为 CSV。
func (s *Service) ExportEvaluationsCSV(stationID string, limit int) ([]byte, error) {
	evs, err := s.store.LatestEvaluations(stationID, limit)
	if err != nil {
		return nil, err
	}
	var sb strings.Builder
	sb.WriteString("station_id,score,danger_level,weak_layer_depth_cm,created_at_utc\n")
	for i := len(evs) - 1; i >= 0; i-- {
		ev := evs[i]
		sb.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s\n",
			ev.StationID, formatFloat(ev.Score), string(ev.DangerLevel),
			formatFloat(ev.WeakLayerDepthCm), ev.CreatedAt.UTC().Format(time.RFC3339)))
	}
	return []byte(sb.String()), nil
}

// DailySummaryCSV 把日统计行渲染为 CSV。
func (s *Service) DailySummaryCSV(rows []DailySummaryRow) []byte {
	var sb strings.Builder
	sb.WriteString("day,station_id,readings,alerts\n")
	for _, row := range rows {
		sb.WriteString(fmt.Sprintf("%s,%s,%d,%d\n", row.Day, row.StationID, row.Readings, row.Alerts))
	}
	return []byte(sb.String())
}

// ParseExportWindow 解析导出窗口参数；缺省最近 24 小时。
func ParseExportWindow(fromStr, toStr string, now time.Time) (time.Time, time.Time, error) {
	to := now
	from := now.Add(-24 * time.Hour)
	var err error
	if toStr != "" {
		if to, err = time.Parse(time.RFC3339, toStr); err != nil {
			return time.Time{}, time.Time{}, model.ErrReadingBadTime
		}
	}
	if fromStr != "" {
		if from, err = time.Parse(time.RFC3339, fromStr); err != nil {
			return time.Time{}, time.Time{}, model.ErrReadingBadTime
		}
	}
	if from.After(to) {
		from, to = to, from
	}
	return from.UTC(), to.UTC(), nil
}
