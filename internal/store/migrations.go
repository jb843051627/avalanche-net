package store

// migrate 建立业务实体表结构（幂等）。
func (s *Store) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS regions (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS stations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			region_id TEXT NOT NULL REFERENCES regions(id),
			elevation_m REAL NOT NULL,
			aspect TEXT NOT NULL,
			lat REAL NOT NULL,
			lon REAL NOT NULL,
			status TEXT NOT NULL DEFAULT 'offline',
			slope_angle_deg REAL NOT NULL DEFAULT 0,
			installed_at TEXT NOT NULL,
			last_heartbeat TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_stations_region ON stations(region_id)`,
		`CREATE TABLE IF NOT EXISTS sensors (
			id TEXT PRIMARY KEY,
			station_id TEXT NOT NULL REFERENCES stations(id),
			kind TEXT NOT NULL,
			unit TEXT NOT NULL,
			warn_at REAL NOT NULL,
			crit_at REAL NOT NULL,
			calibrated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS readings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			station_id TEXT NOT NULL,
			sensor_kind TEXT NOT NULL,
			value REAL NOT NULL,
			recorded_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_readings_station_time ON readings(station_id, recorded_at)`,
		`CREATE TABLE IF NOT EXISTS snow_profiles (
			id TEXT PRIMARY KEY,
			station_id TEXT NOT NULL,
			observed_at TEXT NOT NULL,
			observer TEXT NOT NULL,
			total_cm REAL NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS snow_layers (
			profile_id TEXT NOT NULL REFERENCES snow_profiles(id),
			idx INTEGER NOT NULL,
			depth_from_cm REAL NOT NULL,
			depth_to_cm REAL NOT NULL,
			density_kgm3 REAL NOT NULL,
			grain_shape TEXT NOT NULL,
			hardness TEXT NOT NULL,
			temp_c REAL NOT NULL,
			PRIMARY KEY(profile_id, idx)
		)`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id TEXT PRIMARY KEY,
			station_id TEXT NOT NULL,
			rule_key TEXT NOT NULL,
			level TEXT NOT NULL,
			state TEXT NOT NULL,
			reason TEXT NOT NULL,
			value REAL NOT NULL,
			triggered_at TEXT NOT NULL,
			acked_by TEXT DEFAULT '',
			acked_at TEXT DEFAULT NULL,
			resolved_at TEXT DEFAULT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_station ON alerts(station_id, triggered_at)`,
		`CREATE TABLE IF NOT EXISTS evaluations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			station_id TEXT NOT NULL,
			score REAL NOT NULL,
			danger_level TEXT NOT NULL,
			weak_layer_idx INTEGER NOT NULL,
			weak_layer_depth_cm REAL NOT NULL,
			wind_factor REAL NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS inspections (
			id TEXT PRIMARY KEY,
			station_id TEXT NOT NULL,
			due_date TEXT NOT NULL,
			status TEXT NOT NULL,
			assignee TEXT DEFAULT '',
			notes TEXT DEFAULT '',
			created_at TEXT NOT NULL,
			completed_at TEXT DEFAULT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS bulletins (
			id TEXT PRIMARY KEY,
			region_id TEXT NOT NULL,
			issued_for TEXT NOT NULL,
			stage TEXT NOT NULL,
			above_treeline TEXT NOT NULL,
			near_treeline TEXT NOT NULL,
			below_treeline TEXT NOT NULL,
			summary TEXT NOT NULL,
			created_at TEXT NOT NULL,
			published_at TEXT DEFAULT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS weather_samples (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			region_id TEXT NOT NULL,
			recorded_at TEXT NOT NULL,
			wind_speed_kmh REAL NOT NULL,
			air_temp_c REAL NOT NULL,
			precip_24h_mm REAL NOT NULL,
			new_snow_24h_cm REAL NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_weather_region_time ON weather_samples(region_id, recorded_at)`,
		`CREATE TABLE IF NOT EXISTS calibrations (
			id TEXT PRIMARY KEY,
			station_id TEXT NOT NULL,
			sensor_kind TEXT NOT NULL,
			reference_in REAL NOT NULL,
			reported_out REAL NOT NULL,
			offset_val REAL NOT NULL,
			drift_pct REAL NOT NULL,
			calibrated_by TEXT DEFAULT '',
			calibrated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS observations (
			id TEXT PRIMARY KEY,
			region_id TEXT NOT NULL,
			station_id TEXT DEFAULT '',
			observed_at TEXT NOT NULL,
			aspect TEXT NOT NULL,
			elevation_m REAL NOT NULL,
			type TEXT NOT NULL,
			size INTEGER NOT NULL,
			trigger TEXT NOT NULL,
			reporter TEXT NOT NULL,
			comment TEXT DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_observations_region ON observations(region_id, observed_at)`,
		`CREATE TABLE IF NOT EXISTS avalanche_paths (
			id TEXT PRIMARY KEY,
			region_id TEXT NOT NULL,
			name TEXT NOT NULL,
			start_elev_m REAL NOT NULL,
			end_elev_m REAL NOT NULL,
			aspect TEXT NOT NULL,
			slope_deg REAL NOT NULL,
			length_m REAL NOT NULL,
			hits_road INTEGER NOT NULL DEFAULT 0,
			hits_structs INTEGER NOT NULL DEFAULT 0,
			last_event_at TEXT DEFAULT NULL,
			event_count_12m INTEGER NOT NULL DEFAULT 0,
			registered_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_paths_region ON avalanche_paths(region_id)`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
