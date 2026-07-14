package db

import (
	"context"
	"fmt"
	"github.com/guntisdev/entlite/examples/05-sqlite-optional/ent/logic"
	"time"
	internal "github.com/guntisdev/entlite/examples/05-sqlite-optional/ent/gen/db/internal"
)

type CreateArticleParams struct {
	Slug string `json:"slug"`
	Title string `json:"title"`
	Author string `json:"author"`
	Subtitle *string `json:"subtitle"`
	ReadingMinutes *int32 `json:"reading_minutes"`
	LastViewedMs *int64 `json:"last_viewed_ms"`
	Rating *float64 `json:"rating"`
	CoverImage *[]byte `json:"cover_image"`
	PublishedAt *time.Time `json:"published_at"`
	IsFeatured *bool `json:"is_featured"`
}

func (q *Queries) CreateArticle(ctx context.Context, arg CreateArticleParams) (string, error) {
	if !logic.NotBlank(arg.Title) {
		return "", fmt.Errorf("Failed create: incorrect value for 'Article' in field 'title', validated by 'logic.NotBlank'")
	}
	internalArg := internal.CreateArticleParams{
		ID: logic.NewUUID(),
		Slug: arg.Slug,
		Title: arg.Title,
		Author: arg.Author,
		Subtitle: arg.Subtitle,
		ReadingMinutes: IntPtrConvert[int32, int64](arg.ReadingMinutes),
		LastViewedMs: arg.LastViewedMs,
		Rating: arg.Rating,
		CoverImage: arg.CoverImage,
		PublishedAt: arg.PublishedAt,
		IsFeatured: SQLiteBoolToInt(OptionalWithFallback(arg.IsFeatured, false)),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return (*internal.Queries)(q).CreateArticle(ctx, internalArg)
}

func (q *Queries) DeleteArticle(ctx context.Context, id string) error {
	return (*internal.Queries)(q).DeleteArticle(ctx, id)
}

func (q *Queries) GetArticleByID(ctx context.Context, id string) (*Article, error) {
	dbResult, err := (*internal.Queries)(q).GetArticleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ArticleFromSQL(&dbResult), nil
}

func (q *Queries) GetArticleBySlug(ctx context.Context, slug string) (*Article, error) {
	dbResult, err := (*internal.Queries)(q).GetArticleBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return ArticleFromSQL(&dbResult), nil
}

func (q *Queries) ListArticleByAuthor(ctx context.Context, author string) ([]*Article, error) {
	dbResults, err := (*internal.Queries)(q).ListArticleByAuthor(ctx, author)
	if err != nil {
		return nil, err
	}
	result := make([]*Article, len(dbResults))
	for i := range dbResults {
		result[i] = ArticleFromSQL(&dbResults[i])
	}
	return result, nil
}

type ListArticleFilterByAuthorIsFeaturedPublishedAtTitleParams = internal.ListArticleFilterByAuthorIsFeaturedPublishedAtTitleParams
func (q *Queries) ListArticleFilterByAuthorIsFeaturedPublishedAtTitle(ctx context.Context, arg ListArticleFilterByAuthorIsFeaturedPublishedAtTitleParams) ([]*Article, error) {
	dbResults, err := (*internal.Queries)(q).ListArticleFilterByAuthorIsFeaturedPublishedAtTitle(ctx, arg)
	if err != nil {
		return nil, err
	}
	result := make([]*Article, len(dbResults))
	for i := range dbResults {
		result[i] = ArticleFromSQL(&dbResults[i])
	}
	return result, nil
}

type UpdateArticleParams struct {
	Slug string `json:"slug"`
	Title string `json:"title"`
	Author string `json:"author"`
	Subtitle *string `json:"subtitle"`
	ReadingMinutes *int32 `json:"reading_minutes"`
	LastViewedMs *int64 `json:"last_viewed_ms"`
	Rating *float64 `json:"rating"`
	CoverImage *[]byte `json:"cover_image"`
	PublishedAt *time.Time `json:"published_at"`
	IsFeatured *bool `json:"is_featured"`
}

func (q *Queries) UpdateArticle(ctx context.Context, arg UpdateArticleParams) (*Article, error) {
	if !logic.NotBlank(arg.Title) {
		return nil, fmt.Errorf("Failed update: incorrect value for 'Article' in field 'title', validated by 'logic.NotBlank'")
	}
	internalArg := internal.UpdateArticleParams{
		ID: logic.NewUUID(),
		Slug: arg.Slug,
		Title: arg.Title,
		Author: arg.Author,
		Subtitle: arg.Subtitle,
		ReadingMinutes: IntPtrConvert[int32, int64](arg.ReadingMinutes),
		LastViewedMs: arg.LastViewedMs,
		Rating: arg.Rating,
		CoverImage: arg.CoverImage,
		PublishedAt: arg.PublishedAt,
		IsFeatured: SQLiteBoolPtrToInt64Ptr(arg.IsFeatured),
		UpdatedAt: time.Now(),
	}

	dbArticle, err := (*internal.Queries)(q).UpdateArticle(ctx, internalArg)
	if err != nil {
		return nil, err
	}
	return ArticleFromSQL(&dbArticle), nil
}

