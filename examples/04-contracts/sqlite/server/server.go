package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/guntisdev/entlite/examples/04-contracts/sqlite/ent/gen/db"
	"github.com/guntisdev/entlite/examples/04-contracts/sqlite/ent/gen/pb"
)

type MatchServer struct {
	db *sql.DB
}

// enforces implementation of proto methods
var _ pb.MatchServiceHandler = (*MatchServer)(nil)

func NewMatchServiceServer(db *sql.DB) *MatchServer {
	return &MatchServer{
		db: db,
	}
}

func (s *MatchServer) CreateMatch(
	ctx context.Context,
	req *connect.Request[pb.CreateMatchRequest],
) (*connect.Response[pb.Match], error) {
	log.Printf("Create match: %+v", req.Msg)

	queries := db.New(s.db)

	matchID, err := queries.CreateMatch(ctx, db.CreateMatchParams{
		White:    req.Msg.White,
		Black:    req.Msg.Black,
		Result:   req.Msg.Result,
		Opening:  req.Msg.Opening,
		Moves:    req.Msg.Moves,
		PlayedAt: protoToTimePtr(req.Msg.PlayedAt),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to create match: %w", err))
	}

	match, err := queries.GetMatchByID(ctx, matchID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get created match: %w", err))
	}

	// Audit has no proto contract, this row never leaves the server
	audit(ctx, queries, "created", matchID, describeMatch(match))

	return connect.NewResponse(match.ToProto()), nil
}

func (s *MatchServer) GetMatchByID(
	ctx context.Context,
	req *connect.Request[pb.GetMatchByIDRequest],
) (*connect.Response[pb.Match], error) {
	log.Printf("Get match: ID=%d", req.Msg.ID)

	queries := db.New(s.db)

	match, err := queries.GetMatchByID(ctx, req.Msg.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("match not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get match: %w", err))
	}

	return connect.NewResponse(match.ToProto()), nil
}

func (s *MatchServer) DeleteMatch(
	ctx context.Context,
	req *connect.Request[pb.DeleteMatchRequest],
) (*connect.Response[emptypb.Empty], error) {
	log.Printf("Delete match: ID=%d", req.Msg.ID)

	queries := db.New(s.db)

	if err := queries.DeleteMatch(ctx, req.Msg.ID); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete match: %w", err))
	}

	audit(ctx, queries, "deleted", req.Msg.ID, "removed from the club records")

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *MatchServer) ListAllMatch(
	ctx context.Context,
	req *connect.Request[pb.ListAllMatchRequest],
) (*connect.Response[pb.ListAllMatchResponse], error) {
	log.Printf("List all matches")

	queries := db.New(s.db)

	matches, err := queries.ListAllMatch(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list matches: %w", err))
	}

	protoMatches := make([]*pb.Match, len(matches))
	for i, match := range matches {
		protoMatches[i] = match.ToProto()
	}

	return connect.NewResponse(&pb.ListAllMatchResponse{
		Rows: protoMatches,
	}), nil
}

// AuditCount reports how many audit rows are stored, used by the /audit endpoint
func AuditCount(ctx context.Context, database *sql.DB) (int, error) {
	entries, err := db.New(database).ListAllAudit(ctx)
	if err != nil {
		return 0, err
	}

	return len(entries), nil
}

// audit appends an internal row, a failure here must not fail the rpc
func audit(ctx context.Context, queries *db.Queries, action string, matchID int32, detail string) {
	_, err := queries.CreateAudit(ctx, db.CreateAuditParams{
		Action:    action,
		MatchID:   matchID,
		Detail:    detail,
		CreatedAt: nil,
	})
	if err != nil {
		log.Printf("Failed to write audit entry: %v", err)
	}
}

func describeMatch(match *db.Match) string {
	return fmt.Sprintf("%s vs %s %s in %d moves", match.White, match.Black, match.Result, match.Moves)
}

// protoToTimePtr keeps an unset timestamp as nil, so the default func applies
func protoToTimePtr(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	value := t.AsTime()
	return &value
}

type PlayerServer struct {
	db *sql.DB
}

// enforces implementation of proto methods
var _ pb.PlayerServiceHandler = (*PlayerServer)(nil)

func NewPlayerServiceServer(db *sql.DB) *PlayerServer {
	return &PlayerServer{
		db: db,
	}
}

// GetByName reads one roster entry, index.Primary("name") makes the name the key,
// and the read only contract gives no write rpc
func (s *PlayerServer) GetPlayerByName(
	ctx context.Context,
	req *connect.Request[pb.GetPlayerByNameRequest],
) (*connect.Response[pb.Player], error) {
	log.Printf("Get player: name=%q", req.Msg.Name)

	player, err := db.New(s.db).GetPlayerByName(ctx, req.Msg.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("player not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get player: %w", err))
	}

	return connect.NewResponse(player.ToProto()), nil
}

func (s *PlayerServer) ListAllPlayer(
	ctx context.Context,
	req *connect.Request[pb.ListAllPlayerRequest],
) (*connect.Response[pb.ListAllPlayerResponse], error) {
	log.Printf("List roster")

	players, err := db.New(s.db).ListAllPlayer(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list players: %w", err))
	}

	protoPlayers := make([]*pb.Player, len(players))
	for i, player := range players {
		protoPlayers[i] = player.ToProto()
	}

	return connect.NewResponse(&pb.ListAllPlayerResponse{
		Rows: protoPlayers,
	}), nil
}

// SeedRoster fills the roster once, CreatePlayer is sqlc only
func SeedRoster(ctx context.Context, database *sql.DB) error {
	queries := db.New(database)

	players, err := queries.ListAllPlayer(ctx)
	if err != nil {
		return fmt.Errorf("failed to read roster: %w", err)
	}
	if len(players) > 0 {
		return nil
	}

	roster := []db.CreatePlayerParams{
		{Name: "Vera Menchik", Rating: 2300, Title: playerTitle("WGM")},
		{Name: "Mikhail Tal", Rating: 2700, Title: playerTitle("GM")},
		{Name: "Judit Polgar", Rating: 2735, Title: playerTitle("GM")},
		{Name: "Club Newcomer", Rating: 1400, Title: nil},
	}

	for _, player := range roster {
		if err := queries.CreatePlayer(ctx, player); err != nil {
			return fmt.Errorf("failed to seed player %q: %w", player.Name, err)
		}
	}

	log.Printf("Seeded roster with %d players", len(roster))
	return nil
}

func playerTitle(value string) *string {
	return &value
}

type StandingServer struct {
	db *sql.DB
}

// enforces implementation of proto methods
var _ pb.StandingServiceHandler = (*StandingServer)(nil)

func NewStandingServiceServer(db *sql.DB) *StandingServer {
	return &StandingServer{
		db: db,
	}
}

// ListAll counts the standings from matches, Standing has no table to read from
func (s *StandingServer) ListAllStanding(
	ctx context.Context,
	req *connect.Request[pb.ListAllStandingRequest],
) (*connect.Response[pb.ListAllStandingResponse], error) {
	log.Printf("List standings")

	matches, err := db.New(s.db).ListAllMatch(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list matches: %w", err))
	}

	return connect.NewResponse(&pb.ListAllStandingResponse{
		Rows: countStandings(matches),
	}), nil
}

// countStandings turns match rows into a ranked table, one entry per player
func countStandings(matches []*db.Match) []*pb.Standing {
	byPlayer := make(map[string]*pb.Standing)

	entryFor := func(player string) *pb.Standing {
		if entry, ok := byPlayer[player]; ok {
			return entry
		}
		entry := &pb.Standing{
			ID:     0,
			Player: player,
			Played: 0,
			Wins:   0,
			Draws:  0,
			Losses: 0,
			Points: 0,
		}
		byPlayer[player] = entry
		return entry
	}

	for _, match := range matches {
		white := entryFor(match.White)
		black := entryFor(match.Black)

		white.Played++
		black.Played++

		switch match.Result {
		case "1-0":
			white.Wins++
			white.Points += 1
			black.Losses++
		case "0-1":
			black.Wins++
			black.Points += 1
			white.Losses++
		case "1/2-1/2":
			white.Draws++
			white.Points += 0.5
			black.Draws++
			black.Points += 0.5
		}
	}

	standings := make([]*pb.Standing, 0, len(byPlayer))
	for _, entry := range byPlayer {
		standings = append(standings, entry)
	}

	sort.Slice(standings, func(i, j int) bool {
		if standings[i].Points != standings[j].Points {
			return standings[i].Points > standings[j].Points
		}
		return standings[i].Player < standings[j].Player
	})

	// the id field carries the rank
	for i, entry := range standings {
		entry.ID = int32(i + 1)
	}

	return standings
}
