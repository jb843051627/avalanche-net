package service

import (
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// RegisterPath 登记雪崩路径。
func (s *Service) RegisterPath(p *model.AvalanchePath) error {
	if err := p.Validate(); err != nil {
		return err
	}
	p.RegisteredAt = s.clk.Now().UTC()
	s.met.Inc("path.registered")
	return s.store.InsertPath(p)
}

// GetPath 查询路径。
func (s *Service) GetPath(id string) (*model.AvalanchePath, error) {
	return s.store.GetPath(id)
}

// ListPaths 区域路径列表。
func (s *Service) ListPaths(regionID string) ([]*model.AvalanchePath, error) {
	return s.store.ListPathsByRegion(regionID)
}

// MarkPathEvent 把目击事件关联到路径（更新活跃计数）。
func (s *Service) MarkPathEvent(pathID string, at time.Time) error {
	if _, err := s.store.GetPath(pathID); err != nil {
		return err
	}
	return s.store.MarkPathEvent(pathID, at.UTC())
}

// PathExposureRow 是路径暴露度排行行。
type PathExposureRow struct {
	PathID        string `json:"path_id"`
	Name          string `json:"name"`
	ExposureScore int    `json:"exposure_score"`
	DangerLevel   string `json:"danger_level"`
}

// PathExposureRanking 计算区域路径暴露度排行：
// 基础暴露分（道路/建筑）+ 活跃度 + 当前坡向危险等级加成。
func (s *Service) PathExposureRanking(regionID string) ([]PathExposureRow, error) {
	paths, err := s.store.ListPathsByRegion(regionID)
	if err != nil {
		return nil, err
	}
	cells, err := s.BuildRegionRose(regionID)
	if err != nil {
		return nil, err
	}
	levelByAspect := map[model.Aspect]model.DangerLevel{}
	for _, c := range cells {
		a := model.Aspect(c.Aspect)
		if c.DangerLevel.Rank() > levelByAspect[a].Rank() {
			levelByAspect[a] = c.DangerLevel
		}
	}
	out := make([]PathExposureRow, 0, len(paths))
	for _, p := range paths {
		score := p.ExposureBase() + p.ActivityBonus()
		lv := levelByAspect[p.Aspect]
		score += (lv.Rank() - 1) * 10
		out = append(out, PathExposureRow{
			PathID:        p.ID,
			Name:          p.Name,
			ExposureScore: score,
			DangerLevel:   string(lv),
		})
	}
	sortPathRows(out)
	return out, nil
}

func sortPathRows(rows []PathExposureRow) {
	for i := 1; i < len(rows); i++ {
		for j := i; j > 0 && rows[j].ExposureScore > rows[j-1].ExposureScore; j-- {
			rows[j], rows[j-1] = rows[j-1], rows[j]
		}
	}
}
