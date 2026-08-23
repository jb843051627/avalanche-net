package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jb843051627/avalanche-net/internal/model"
)

// InsertBulletin 落库公报草稿。
func (s *Store) InsertBulletin(b *model.Bulletin) error {
	_, err := s.db.Exec(`INSERT INTO bulletins(id,region_id,issued_for,stage,above_treeline,near_treeline,below_treeline,summary,created_at)
		VALUES(?,?,?,?,?,?,?,?,?)`,
		b.ID, b.RegionID, fmtTime(b.IssuedFor), string(b.Stage),
		string(b.AboveTreeline), string(b.NearTreeline), string(b.BelowTreeline),
		b.Summary, fmtTime(b.CreatedAt))
	return err
}

const bulletinCols = `id,region_id,issued_for,stage,above_treeline,near_treeline,below_treeline,summary,created_at,published_at`

func scanBulletin(row interface{ Scan(...any) error }) (*model.Bulletin, error) {
	var b model.Bulletin
	var stage, issued, created string
	var published sql.NullString
	err := row.Scan(&b.ID, &b.RegionID, &issued, &stage,
		&b.AboveTreeline, &b.NearTreeline, &b.BelowTreeline, &b.Summary, &created, &published)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrBulletinNotFound
	}
	if err != nil {
		return nil, err
	}
	b.Stage = model.BulletinStage(stage)
	b.IssuedFor = parseTime(issued)
	b.CreatedAt = parseTime(created)
	b.PublishedAt = nullableTime(published)
	return &b, nil
}

// GetBulletin 按 ID 查询公报。
func (s *Store) GetBulletin(id string) (*model.Bulletin, error) {
	return scanBulletin(s.db.QueryRow(`SELECT `+bulletinCols+` FROM bulletins WHERE id=?`, id))
}

// UpdateBulletinStage 推进公报阶段并写发布时间。
func (s *Store) UpdateBulletinStage(id string, stage model.BulletinStage, publishedAt *time.Time) error {
	var pub any
	if publishedAt != nil {
		pub = fmtTime(*publishedAt)
	}
	res, err := s.db.Exec(`UPDATE bulletins SET stage=?, published_at=COALESCE(?,published_at) WHERE id=?`,
		string(stage), pub, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrBulletinNotFound
	}
	return nil
}

// ListPublishedBulletins 返回某区域已发布的公报（含归档），时间倒序。
func (s *Store) ListPublishedBulletins(regionID string) ([]*model.Bulletin, error) {
	rows, err := s.db.Query(`SELECT `+bulletinCols+` FROM bulletins
		WHERE region_id=? AND stage IN ('published','archived') ORDER BY issued_for DESC LIMIT 200`, regionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Bulletin
	for rows.Next() {
		b, err := scanBulletin(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// CountBulletinsByStage 统计各阶段公报数量。
func (s *Store) CountBulletinsByStage(stage model.BulletinStage) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM bulletins WHERE stage=?`, string(stage)).Scan(&n)
	return n, err
}
