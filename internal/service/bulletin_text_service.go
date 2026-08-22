package service

import (
	"time"

	"github.com/jb843051627/avalanche-net/internal/engine"
	"github.com/jb843051627/avalanche-net/internal/model"
)

// BuildBulletinText 为区域生成公报正文（玫瑰图 + 24h 加载概况）。
func (s *Service) BuildBulletinText(regionID string, issuedFor time.Time) (*engine.BulletinText, error) {
	cells, err := s.BuildRegionRose(regionID)
	if err != nil {
		return nil, err
	}
	summary, err := s.RegionLoadingSummary(regionID, 24*time.Hour)
	if err != nil {
		return nil, err
	}
	loading := &engine.LoadingSummaryInput{
		MaxWindKmh:   summary.MaxWindKmh,
		NewSnow24hCm: summary.NewSnow24hCm,
		RapidLoading: summary.RapidLoading,
	}
	text := engine.GenerateBulletinText(regionID, issuedFor.UTC(), cells, loading)
	s.met.Inc("bulletin.text_generated")
	return &text, nil
}

// AutoDraftBulletin 由玫瑰图与加载概况自动生成公报草稿并落库。
func (s *Service) AutoDraftBulletin(regionID string, issuedFor time.Time) (*model.Bulletin, error) {
	if regionID == "" {
		return nil, model.ErrInvalidStation
	}
	cells, err := s.BuildRegionRose(regionID)
	if err != nil {
		return nil, err
	}
	text, err := s.BuildBulletinText(regionID, issuedFor)
	if err != nil {
		return nil, err
	}
	b := &model.Bulletin{
		ID:            "bul-" + regionID + "-" + issuedFor.UTC().Format("20060102"),
		RegionID:      regionID,
		IssuedFor:     issuedFor.UTC(),
		AboveTreeline: engine.MaxBandLevel(cells, "above"),
		NearTreeline:  engine.MaxBandLevel(cells, "near"),
		BelowTreeline: engine.MaxBandLevel(cells, "below"),
		Summary:       text.Summary,
	}
	if err := b.Validate(); err != nil {
		return nil, err
	}
	b.Stage = model.BulletinDraft
	b.CreatedAt = s.clk.Now().UTC()
	s.met.Inc("bulletin.auto_drafted")
	return b, s.store.InsertBulletin(b)
}

// WeeklyTrendRow 是站点周趋势行。
type WeeklyTrendRow struct {
	Day       string  `json:"day"`
	ScoreAvg  float64 `json:"score_avg"`
	EvalCount int     `json:"eval_count"`
}

// StationWeeklyTrend 汇总站点最近 n 天的评估均值趋势。
func (s *Service) StationWeeklyTrend(stationID string, days int) ([]WeeklyTrendRow, error) {
	evs, err := s.store.LatestEvaluations(stationID, days*24)
	if err != nil {
		return nil, err
	}
	index := map[string]*WeeklyTrendRow{}
	var order []string
	for i := len(evs) - 1; i >= 0; i-- { // 时间升序遍历
		day := evs[i].CreatedAt.UTC().Format("2006-01-02")
		row := index[day]
		if row == nil {
			row = &WeeklyTrendRow{Day: day}
			index[day] = row
			order = append(order, day)
		}
		row.ScoreAvg += evs[i].Score
		row.EvalCount++
	}
	out := make([]WeeklyTrendRow, 0, len(order))
	for _, day := range order {
		row := index[day]
		row.ScoreAvg = round2(row.ScoreAvg / float64(row.EvalCount))
		out = append(out, *row)
	}
	return out, nil
}
