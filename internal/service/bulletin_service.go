package service

import (
	"github.com/jb843051627/avalanche-net/internal/engine"
	"github.com/jb843051627/avalanche-net/internal/model"
)

// DraftBulletin 创建公报草稿（校验三带等级单调性）。
func (s *Service) DraftBulletin(b *model.Bulletin) error {
	if err := b.Validate(); err != nil {
		return err
	}
	b.Stage = model.BulletinDraft
	b.CreatedAt = s.clk.Now().UTC()
	_ = s.store.InsertBulletin(b)
	s.met.Inc("bulletin.drafted")
	return nil
}

// GetBulletin 查询公报。
func (s *Service) GetBulletin(id string) (*model.Bulletin, error) {
	return s.store.GetBulletin(id)
}

// PublishBulletin 发布公报：draft -> published。
func (s *Service) PublishBulletin(id string) (*model.Bulletin, error) {
	b, err := s.store.GetBulletin(id)
	if err != nil {
		return nil, err
	}
	if !b.Stage.CanTransition(model.BulletinPublished) {
		return nil, model.ErrBulletinStage
	}
	now := s.clk.Now().UTC()
	if err := s.store.UpdateBulletinStage(id, model.BulletinPublished, &now); err != nil {
		return nil, err
	}
	s.met.Inc("bulletin.published")
	return s.store.GetBulletin(id)
}

// ArchiveBulletin 归档公报：published -> archived；draft 不能直接归档。
func (s *Service) ArchiveBulletin(id string) (*model.Bulletin, error) {
	b, err := s.store.GetBulletin(id)
	if err != nil {
		return nil, err
	}
	if !b.Stage.CanTransition(model.BulletinArchived) {
		return nil, model.ErrBulletinStage
	}
	if err := s.store.UpdateBulletinStage(id, model.BulletinArchived, nil); err != nil {
		return nil, err
	}
	s.met.Inc("bulletin.archived")
	return s.store.GetBulletin(id)
}

// ListPublishedBulletins 返回区域已发布公报。
func (s *Service) ListPublishedBulletins(regionID string) ([]*model.Bulletin, error) {
	return s.store.ListPublishedBulletins(regionID)
}

// SuggestLevels 依据区域内站点最新评估，给出三带危险等级建议值。
// 取各海拔带内最高的评估等级作为建议。
func (s *Service) SuggestLevels(regionID string) (above, near, below model.DangerLevel, err error) {
	stations, err := s.store.ListStations(regionID)
	if err != nil {
		return "", "", "", err
	}
	above = model.DangerLow
	near = model.DangerLow
	below = model.DangerLow
	for _, st := range stations {
		evs, err := s.store.LatestEvaluations(st.ID, 1)
		if err != nil || len(evs) == 0 {
			continue
		}
		lv := evs[0].DangerLevel
		switch engine.ElevationBand(st.ElevationM) {
		case "above":
			if lv.Rank() > above.Rank() {
				above = lv
			}
		case "near":
			if lv.Rank() > near.Rank() {
				near = lv
			}
		default:
			if lv.Rank() > below.Rank() {
				below = lv
			}
		}
	}
	return above, near, below, nil
}
