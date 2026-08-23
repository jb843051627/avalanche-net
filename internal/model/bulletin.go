package model

import (
	"errors"
	"time"
)

var (
	ErrBulletinNotFound = errors.New("bulletin not found")
	ErrBulletinStage    = errors.New("invalid bulletin stage transition")
)

// BulletinStage 公报阶段机：draft -> published -> archived。
type BulletinStage string

const (
	BulletinDraft     BulletinStage = "draft"
	BulletinPublished BulletinStage = "published"
	BulletinArchived  BulletinStage = "archived"
)

// CanTransition 校验公报阶段迁移。draft 不能直接归档。
func (s BulletinStage) CanTransition(to BulletinStage) bool {
	switch s {
	case BulletinDraft:
		return to == BulletinPublished
	case BulletinPublished:
		return to == BulletinArchived
	case BulletinArchived:
		return false
	}
	return false
}

// Bulletin 是一份区域雪崩预报公报，按三个海拔带分别给出危险等级。
type Bulletin struct {
	ID            string        `json:"id"`
	RegionID      string        `json:"region_id"`
	IssuedFor     time.Time     `json:"issued_for"`
	Stage         BulletinStage `json:"stage"`
	AboveTreeline DangerLevel   `json:"above_treeline"`
	NearTreeline  DangerLevel   `json:"near_treeline"`
	BelowTreeline DangerLevel   `json:"below_treeline"`
	Summary       string        `json:"summary"`
	CreatedAt     time.Time     `json:"created_at"`
	PublishedAt   *time.Time    `json:"published_at,omitempty"`
}

// Validate 校验公报字段：三带等级齐全且上林线 >= 近林线 >= 下林线。
func (b *Bulletin) Validate() error {
	if b.ID == "" || b.RegionID == "" {
		return ErrInvalidStation
	}
	if b.AboveTreeline.Rank() < b.NearTreeline.Rank() ||
		b.NearTreeline.Rank() < b.BelowTreeline.Rank() {
		return ErrInvalidDanger
	}
	if b.AboveTreeline.Rank() == 0 || b.NearTreeline.Rank() == 0 || b.BelowTreeline.Rank() == 0 {
		return ErrInvalidDanger
	}
	return nil
}
