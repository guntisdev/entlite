package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/guntisdev/entlite/examples/03-optional/sqlite/ent/gen/db"
	"github.com/guntisdev/entlite/examples/03-optional/sqlite/ent/gen/pb"
)

type ArticleServer struct {
	db *sql.DB
}

// enforces implementation of proto methods
var _ pb.ArticleServiceHandler = (*ArticleServer)(nil)

func NewArticleServiceServer(db *sql.DB) *ArticleServer {
	return &ArticleServer{
		db: db,
	}
}

func (s *ArticleServer) Create(
	ctx context.Context,
	req *connect.Request[pb.CreateArticleRequest],
) (*connect.Response[pb.Article], error) {
	log.Printf("Create article: %+v", req.Msg)

	queries := db.New(s.db)

	// optional fields stay pointers, nil writes NULL
	articleID, err := queries.CreateArticle(ctx, db.CreateArticleParams{
		Slug:           req.Msg.Slug,
		Title:          req.Msg.Title,
		Author:         req.Msg.Author,
		Subtitle:       req.Msg.Subtitle,
		ReadingMinutes: req.Msg.ReadingMinutes,
		LastViewedMs:   req.Msg.LastViewedMs,
		Rating:         req.Msg.Rating,
		CoverImage:     db.NullBytesToPtr(req.Msg.CoverImage),
		PublishedAt:    protoToTimePtr(req.Msg.PublishedAt),
		Metadata:       req.Msg.Metadata,
		IsFeatured:     req.Msg.IsFeatured,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create article: %w", err))
	}

	article, err := queries.GetArticleByID(ctx, articleID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get created article: %w", err))
	}

	return connect.NewResponse(article.ToProto()), nil
}

func (s *ArticleServer) GetByID(
	ctx context.Context,
	req *connect.Request[pb.GetArticleByIDRequest],
) (*connect.Response[pb.Article], error) {
	log.Printf("Get article: ID=%s", req.Msg.ID)

	queries := db.New(s.db)

	article, err := queries.GetArticleByID(ctx, req.Msg.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("article not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get article: %w", err))
	}

	return connect.NewResponse(article.ToProto()), nil
}

func (s *ArticleServer) GetBySlug(
	ctx context.Context,
	req *connect.Request[pb.GetArticleBySlugRequest],
) (*connect.Response[pb.Article], error) {
	log.Printf("Get article by slug: slug=%s", req.Msg.Slug)

	queries := db.New(s.db)

	article, err := queries.GetArticleBySlug(ctx, req.Msg.Slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("article not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get article by slug: %w", err))
	}

	return connect.NewResponse(article.ToProto()), nil
}

func (s *ArticleServer) Update(
	ctx context.Context,
	req *connect.Request[pb.UpdateArticleRequest],
) (*connect.Response[pb.Article], error) {
	log.Printf("Update article: ID=%s, %+v", req.Msg.GetID(), req.Msg)

	queries := db.New(s.db)

	// TODO db.UpdateArticleParams has no ID field, so req.Msg.ID cannot be passed
	article, err := queries.UpdateArticle(ctx, db.UpdateArticleParams{
		Slug:           req.Msg.Slug,
		Title:          req.Msg.Title,
		Author:         req.Msg.Author,
		Subtitle:       req.Msg.Subtitle,
		ReadingMinutes: req.Msg.ReadingMinutes,
		LastViewedMs:   req.Msg.LastViewedMs,
		Rating:         req.Msg.Rating,
		CoverImage:     db.NullBytesToPtr(req.Msg.CoverImage),
		PublishedAt:    protoToTimePtr(req.Msg.PublishedAt),
		Metadata:       req.Msg.Metadata,
		IsFeatured:     req.Msg.IsFeatured,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("article not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update article: %w", err))
	}

	return connect.NewResponse(article.ToProto()), nil
}

func (s *ArticleServer) Delete(
	ctx context.Context,
	req *connect.Request[pb.DeleteArticleRequest],
) (*connect.Response[emptypb.Empty], error) {
	log.Printf("Delete article: ID=%s", req.Msg.ID)

	queries := db.New(s.db)

	if err := queries.DeleteArticle(ctx, req.Msg.ID); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete article: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *ArticleServer) ListAll(
	ctx context.Context,
	req *connect.Request[pb.ListAllArticleRequest],
) (*connect.Response[pb.ListAllArticleResponse], error) {
	log.Printf("List all articles")

	queries := db.New(s.db)

	dbArticles, err := queries.ListAllArticle(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list all articles: %w", err))
	}

	return connect.NewResponse(&pb.ListAllArticleResponse{
		Articles: toProtoArticles(dbArticles),
	}), nil
}

func (s *ArticleServer) ListByAuthor(
	ctx context.Context,
	req *connect.Request[pb.ListArticleByAuthorRequest],
) (*connect.Response[pb.ListArticleByAuthorResponse], error) {
	log.Printf("List articles by author: author=%s", req.Msg.Author)

	queries := db.New(s.db)

	dbArticles, err := queries.ListArticleByAuthor(ctx, db.ListArticleByAuthorParams{
		Author: req.Msg.Author,
		Limit:  req.Msg.GetLimit(),
		Offset: req.Msg.GetOffset(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list articles: %w", err))
	}

	return connect.NewResponse(&pb.ListArticleByAuthorResponse{
		Articles: toProtoArticles(dbArticles),
	}), nil
}

func (s *ArticleServer) FilterByAuthorIsFeaturedPublishedAtTitle(
	ctx context.Context,
	req *connect.Request[pb.ListArticleFilterByAuthorIsFeaturedPublishedAtTitleRequest],
) (*connect.Response[pb.ListArticleFilterByAuthorIsFeaturedPublishedAtTitleResponse], error) {
	log.Printf("Filter articles: author=%s, is_featured=%t, title=%s",
		req.Msg.Author, req.Msg.GetIsFeatured(), req.Msg.GetTitle())

	queries := db.New(s.db)

	// TODO generated params miss the optional published_at range
	dbArticles, err := queries.ListArticleFilterByAuthorIsFeaturedPublishedAtTitle(
		ctx,
		db.ListArticleFilterByAuthorIsFeaturedPublishedAtTitleParams{
			Author:     req.Msg.Author,
			IsFeatured: req.Msg.GetIsFeatured(),
			Title:      req.Msg.GetTitle(),
			Limit:      req.Msg.GetLimit(),
			Offset:     req.Msg.GetOffset(),
		},
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to filter articles: %w", err))
	}

	return connect.NewResponse(&pb.ListArticleFilterByAuthorIsFeaturedPublishedAtTitleResponse{
		Articles: toProtoArticles(dbArticles),
	}), nil
}

// protoToTimePtr keeps an unset timestamp as nil, not zero time
func protoToTimePtr(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	value := t.AsTime()
	return &value
}

func toProtoArticles(dbArticles []*db.Article) []*pb.Article {
	articles := make([]*pb.Article, len(dbArticles))
	for i, dbArticle := range dbArticles {
		articles[i] = dbArticle.ToProto()
	}
	return articles
}
