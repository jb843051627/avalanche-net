package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// InsertWeatherSample 落库区域气象样本。
func (s *Store) InsertWeatherSample(w *model.WeatherSample) error {
	_, err := s.db.Exec(`INSERT INTO weather_samples(region_id,recorded_at,wind_speed_kmh,air_temp_c,precip_24h_mm,new_snow_24h_cm)
		VALUES(?,?,?,?,?,?)`,
		w.RegionID, fmtTime(w.RecordedAt), w.WindSpeedKmh, w.AirTempC, w.Precip24hMm, w.NewSnow24hCm)
	return err
}

// LatestWeatherSample 返回区域最新一条气象样本；无记录返回 nil。
func (s *Store) LatestWeatherSample(regionID string) (*model.WeatherSample, error) {
	row := s.db.QueryRow(`SELECT region_id,recorded_at,wind_speed_kmh,air_temp_c,precip_24h_mm,new_snow_24h_cm
		FROM weather_samples WHERE region_id=? ORDER BY recorded_at DESC LIMIT 1`, regionID)
	return scanWeather(row)
}

func scanWeather(row interface{ Scan(...any) error }) (*model.WeatherSample, error) {
	var w model.WeatherSample
	var at string
	err := row.Scan(&w.RegionID, &at, &w.WindSpeedKmh, &w.AirTempC, &w.Precip24hMm, &w.NewSnow24hCm)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	w.RecordedAt = parseTime(at)
	return &w, nil
}

// ListWeatherSince 返回区域自 since 以来的全部气象样本（时间升序）。
func (s *Store) ListWeatherSince(regionID string, since time.Time) ([]model.WeatherSample, error) {
	rows, err := s.db.Query(`SELECT region_id,recorded_at,wind_speed_kmh,air_temp_c,precip_24h_mm,new_snow_24h_cm
		FROM weather_samples WHERE region_id=? AND recorded_at>=? ORDER BY recorded_at`, regionID, fmtTime(since))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.WeatherSample
	for rows.Next() {
		var w model.WeatherSample
		var at string
		if err := rows.Scan(&w.RegionID, &at, &w.WindSpeedKmh, &w.AirTempC, &w.Precip24hMm, &w.NewSnow24hCm); err != nil {
			return nil, err
		}
		w.RecordedAt = parseTime(at)
		out = append(out, w)
	}
	return out, rows.Err()
}
