package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/guntisdev/entlite/examples/05-queries/sqlite/ent/gen/db"
	"github.com/guntisdev/entlite/examples/05-queries/sqlite/ent/gen/pb"
)

type BuildServer struct {
	db *sql.DB
}

// enforces implementation of proto methods
var _ pb.BuildServiceHandler = (*BuildServer)(nil)

func NewBuildServiceServer(db *sql.DB) *BuildServer {
	return &BuildServer{
		db: db,
	}
}

func (s *BuildServer) CreateBuild(
	ctx context.Context,
	req *connect.Request[pb.CreateBuildRequest],
) (*connect.Response[pb.Build], error) {
	log.Printf("Create build: commit_sha=%s", req.Msg.CommitSha)

	queries := db.New(s.db)

	// Upsert("commit_sha") means a re-run returns the id of the row it overwrote
	buildID, err := queries.CreateBuild(ctx, db.CreateBuildParams{
		CommitSha:   req.Msg.CommitSha,
		Branch:      req.Msg.Branch,
		Env:         req.Msg.Env,
		Status:      req.Msg.Status,
		Author:      req.Msg.Author,
		Message:     req.Msg.Message,
		DurationMs:  req.Msg.DurationMs,
		StartedAt:   req.Msg.StartedAt.AsTime(),
		FailedTests: req.Msg.FailedTests,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create build: %w", err))
	}

	build, err := queries.GetBuildByID(ctx, buildID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get created build: %w", err))
	}

	return connect.NewResponse(build.ToProto()), nil
}

func (s *BuildServer) CreateBulkBuild(
	ctx context.Context,
	req *connect.Request[pb.CreateBulkBuildRequest],
) (*connect.Response[pb.CreateBulkBuildResponse], error) {
	log.Printf("Create bulk builds: %d rows", len(req.Msg.Rows))

	queries := db.New(s.db)

	params := make([]db.CreateBulkBuildParams, 0, len(req.Msg.Rows))
	for _, row := range req.Msg.Rows {
		params = append(params, db.CreateBulkBuildParams{
			CommitSha:   row.CommitSha,
			Branch:      row.Branch,
			Env:         row.Env,
			Status:      row.Status,
			Author:      row.Author,
			Message:     row.Message,
			DurationMs:  row.DurationMs,
			StartedAt:   row.StartedAt.AsTime(),
			FailedTests: row.FailedTests,
		})
	}

	buildIDs, err := queries.CreateBulkBuild(ctx, params)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create builds: %w", err))
	}

	builds := make([]*pb.Build, 0, len(buildIDs))
	for _, buildID := range buildIDs {
		build, err := queries.GetBuildByID(ctx, buildID)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get created build: %w", err))
		}
		builds = append(builds, build.ToProto())
	}

	return connect.NewResponse(&pb.CreateBulkBuildResponse{Rows: builds}), nil
}

func (s *BuildServer) GetBuildByID(
	ctx context.Context,
	req *connect.Request[pb.GetBuildByIDRequest],
) (*connect.Response[pb.Build], error) {
	log.Printf("Get build: ID=%d", req.Msg.ID)

	queries := db.New(s.db)

	build, err := queries.GetBuildByID(ctx, req.Msg.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("build not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get build: %w", err))
	}

	return connect.NewResponse(build.ToProto()), nil
}

func (s *BuildServer) GetBuildByCommitSha(
	ctx context.Context,
	req *connect.Request[pb.GetBuildByCommitShaRequest],
) (*connect.Response[pb.Build], error) {
	log.Printf("Get build by commit: commit_sha=%s", req.Msg.CommitSha)

	queries := db.New(s.db)

	build, err := queries.GetBuildByCommitSha(ctx, req.Msg.CommitSha)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("build not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get build by commit: %w", err))
	}

	return connect.NewResponse(build.ToProto()), nil
}

func (s *BuildServer) UpdateBuild(
	ctx context.Context,
	req *connect.Request[pb.UpdateBuildRequest],
) (*connect.Response[pb.Build], error) {
	log.Printf("Update build: ID=%d", req.Msg.ID)

	queries := db.New(s.db)

	build, err := queries.UpdateBuild(ctx, db.UpdateBuildParams{
		ID:          req.Msg.ID,
		CommitSha:   req.Msg.CommitSha,
		Branch:      req.Msg.Branch,
		Env:         req.Msg.Env,
		Status:      req.Msg.Status,
		Author:      req.Msg.Author,
		Message:     req.Msg.Message,
		DurationMs:  req.Msg.DurationMs,
		StartedAt:   req.Msg.StartedAt.AsTime(),
		FailedTests: req.Msg.FailedTests,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("build not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update build: %w", err))
	}

	return connect.NewResponse(build.ToProto()), nil
}

func (s *BuildServer) DeleteBuild(
	ctx context.Context,
	req *connect.Request[pb.DeleteBuildRequest],
) (*connect.Response[emptypb.Empty], error) {
	log.Printf("Delete build: ID=%d", req.Msg.ID)

	queries := db.New(s.db)

	if err := queries.DeleteBuild(ctx, req.Msg.ID); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete build: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *BuildServer) DeleteAllBuild(
	ctx context.Context,
	req *connect.Request[pb.DeleteAllBuildRequest],
) (*connect.Response[emptypb.Empty], error) {
	log.Printf("Delete all builds")

	queries := db.New(s.db)

	if err := queries.DeleteAllBuild(ctx); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete all builds: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *BuildServer) ListAllBuild(
	ctx context.Context,
	req *connect.Request[pb.ListAllBuildRequest],
) (*connect.Response[pb.ListAllBuildResponse], error) {
	log.Printf("List all builds")

	queries := db.New(s.db)

	builds, err := queries.ListAllBuild(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list all builds: %w", err))
	}

	return connect.NewResponse(&pb.ListAllBuildResponse{
		Rows: toProtoBuilds(builds),
	}), nil
}

// one string field, so the generated method takes it as a plain argument
func (s *BuildServer) ListBuildByBranch(
	ctx context.Context,
	req *connect.Request[pb.ListBuildByBranchRequest],
) (*connect.Response[pb.ListBuildByBranchResponse], error) {
	log.Printf("List builds by branch: branch=%s", req.Msg.Branch)

	queries := db.New(s.db)

	builds, err := queries.ListBuildByBranch(ctx, req.Msg.Branch)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list builds: %w", err))
	}

	return connect.NewResponse(&pb.ListBuildByBranchResponse{
		Rows: toProtoBuilds(builds),
	}), nil
}

// two fields, so they arrive as a params struct
func (s *BuildServer) ListBranchStatusHistory(
	ctx context.Context,
	req *connect.Request[pb.ListBranchStatusHistoryRequest],
) (*connect.Response[pb.ListBranchStatusHistoryResponse], error) {
	log.Printf("List branch history: branch=%s, status=%s", req.Msg.Branch, req.Msg.Status)

	queries := db.New(s.db)

	builds, err := queries.ListBranchStatusHistory(ctx, db.ListBranchStatusHistoryParams{
		Branch: req.Msg.Branch,
		Status: req.Msg.Status,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list branch history: %w", err))
	}

	return connect.NewResponse(&pb.ListBranchStatusHistoryResponse{
		Rows: toProtoBuilds(builds),
	}), nil
}

// the row count is fixed in the sql, so the request carries no limit
func (s *BuildServer) ListLatestBuilds(
	ctx context.Context,
	req *connect.Request[pb.ListLatestBuildsRequest],
) (*connect.Response[pb.ListLatestBuildsResponse], error) {
	log.Printf("List latest builds: branch=%s", req.Msg.Branch)

	queries := db.New(s.db)

	builds, err := queries.ListLatestBuilds(ctx, req.Msg.Branch)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list latest builds: %w", err))
	}

	return connect.NewResponse(&pb.ListLatestBuildsResponse{
		Rows: toProtoBuilds(builds),
	}), nil
}

// Count() adds a second return value, the total before Limit
func (s *BuildServer) SearchBuilds(
	ctx context.Context,
	req *connect.Request[pb.SearchBuildsRequest],
) (*connect.Response[pb.SearchBuildsResponse], error) {
	log.Printf("Search builds: env=%s, message=%s, status=%s",
		req.Msg.Env, req.Msg.Message, req.Msg.Status)

	queries := db.New(s.db)

	builds, totalSize, err := queries.SearchBuilds(ctx, db.SearchBuildsParams{
		Env:          req.Msg.Env,
		MinStartedAt: req.Msg.MinStartedAt.AsTime(),
		MaxStartedAt: req.Msg.MaxStartedAt.AsTime(),
		Message:      req.Msg.Message,
		Status:       req.Msg.Status,
		Limit:        req.Msg.GetLimit(),
		Offset:       req.Msg.GetOffset(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to search builds: %w", err))
	}

	return connect.NewResponse(&pb.SearchBuildsResponse{
		Rows:      toProtoBuilds(builds),
		TotalSize: totalSize,
	}), nil
}

// Distinct("branch") returns the values, so there are no entities to convert
func (s *BuildServer) ListBranches(
	ctx context.Context,
	req *connect.Request[pb.ListBranchesRequest],
) (*connect.Response[pb.ListBranchesResponse], error) {
	log.Printf("List distinct branches")

	queries := db.New(s.db)

	branches, err := queries.ListBranches(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list branches: %w", err))
	}

	return connect.NewResponse(&pb.ListBranchesResponse{Branch: branches}), nil
}

func (s *BuildServer) ListEnvBranches(
	ctx context.Context,
	req *connect.Request[pb.ListEnvBranchesRequest],
) (*connect.Response[pb.ListEnvBranchesResponse], error) {
	log.Printf("List distinct branches of one env: env=%s", req.Msg.Env)

	queries := db.New(s.db)

	branches, err := queries.ListEnvBranches(ctx, req.Msg.Env)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list env branches: %w", err))
	}

	return connect.NewResponse(&pb.ListEnvBranchesResponse{Branch: branches}), nil
}

// several distinct columns come back as the query's own row struct
func (s *BuildServer) ListBranchStatuses(
	ctx context.Context,
	req *connect.Request[pb.ListBranchStatusesRequest],
) (*connect.Response[pb.ListBranchStatusesResponse], error) {
	log.Printf("List distinct branch and status pairs")

	queries := db.New(s.db)

	pairs, err := queries.ListBranchStatuses(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list branch statuses: %w", err))
	}

	rows := make([]*pb.ListBranchStatusesRow, len(pairs))
	for i, pair := range pairs {
		rows[i] = &pb.ListBranchStatusesRow{
			Branch: pair.Branch,
			Status: pair.Status,
		}
	}

	return connect.NewResponse(&pb.ListBranchStatusesResponse{Rows: rows}), nil
}

func toProtoBuilds(dbBuilds []*db.Build) []*pb.Build {
	builds := make([]*pb.Build, len(dbBuilds))
	for i, dbBuild := range dbBuilds {
		builds[i] = dbBuild.ToProto()
	}
	return builds
}
