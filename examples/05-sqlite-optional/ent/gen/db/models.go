package db

import (
	pb "github.com/guntisdev/entlite/examples/05-sqlite-optional/ent/gen/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
	internal "github.com/guntisdev/entlite/examples/05-sqlite-optional/ent/gen/db/internal"
)


type Article struct {
	ID string `json:"id"`
	Slug string `json:"slug"`
	Title string `json:"title"`
	Author string `json:"author"`
	Subtitle *string `json:"subtitle"`
	ReadingMinutes *int32 `json:"reading_minutes"`
	LastViewedMs *int64 `json:"last_viewed_ms"`
	Rating *float64 `json:"rating"`
	CoverImage *[]byte `json:"cover_image"`
	PublishedAt *time.Time `json:"published_at"`
	IsFeatured bool `json:"is_featured"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (m *Article) ArticleToSQL() *internal.Article {
	if m == nil {
		return nil
	}

	return &internal.Article{
		ID: m.ID,
		Slug: m.Slug,
		Title: m.Title,
		Author: m.Author,
		Subtitle: m.Subtitle,
		ReadingMinutes: IntPtrConvert[int32, int64](m.ReadingMinutes),
		LastViewedMs: m.LastViewedMs,
		Rating: m.Rating,
		CoverImage: PtrToNullBytes(m.CoverImage),
		PublishedAt: m.PublishedAt,
		IsFeatured: SQLiteBoolToInt(m.IsFeatured),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func ArticleFromSQL(db *internal.Article) *Article {
	if db == nil {
		return nil
	}

	return &Article{
		ID: db.ID,
		Slug: db.Slug,
		Title: db.Title,
		Author: db.Author,
		Subtitle: db.Subtitle,
		ReadingMinutes: IntPtrConvert[int64, int32](db.ReadingMinutes),
		LastViewedMs: db.LastViewedMs,
		Rating: db.Rating,
		CoverImage: NullBytesToPtr(db.CoverImage),
		PublishedAt: db.PublishedAt,
		IsFeatured: SQLiteIntToBool(db.IsFeatured),
		CreatedAt: db.CreatedAt,
		UpdatedAt: db.UpdatedAt,
	}
}

// ToProto converts Article to proto format
func (m *Article) ToProto() *pb.Article {
	if m == nil {
		return nil
	}

	return &pb.Article{
		Id: m.ID,
		Slug: m.Slug,
		Title: m.Title,
		Author: m.Author,
		Subtitle: m.Subtitle,
		ReadingMinutes: m.ReadingMinutes,
		LastViewedMs: m.LastViewedMs,
		Rating: m.Rating,
		CoverImage: PtrToNullBytes(m.CoverImage),
		PublishedAt: func() *timestamppb.Timestamp { if m.PublishedAt != nil { return timestamppb.New(*m.PublishedAt) }; return nil }(),
		IsFeatured: m.IsFeatured,
		CreatedAt: timestamppb.New(m.CreatedAt),
		UpdatedAt: timestamppb.New(m.UpdatedAt),
	}
}

