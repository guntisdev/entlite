import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { timestampDate, timestampFromDate } from "@bufbuild/protobuf/wkt";
import * as schema from "../../ent/gen/ts/schema_pb.js";
import { MatchService, PlayerService, StandingService } from "../../ent/gen/ts/schema_pb.js";
import type { CreateMatchRequest, Match, Player, Standing } from "../../ent/gen/ts/schema_pb.js";
import {
    daysAgo,
    numberInput,
    randomMoves,
    randomOpening,
    randomPlayer,
    randomResult,
    textInput,
    toString,
} from "./utils.js";

type StrictMessageInput<T extends { $typeName: string; $unknown?: unknown }> = Omit<T, "$typeName" | "$unknown">;

const transport = createConnectTransport({
    baseUrl: window.location.origin,
});

const matchClient = createClient(MatchService, transport);
const playerClient = createClient(PlayerService, transport);
const standingClient = createClient(StandingService, transport);

function log(message: string, data?: any) {
    console.log(message, data);
    const output = document.getElementById("output")!;
    const line = document.createElement("div");
    line.textContent = data !== undefined ? `${message} ${toString(data)}` : message;
    output.appendChild(line);
    output.scrollTop = output.scrollHeight;
}

// prefills the id inputs with the last created ID
function rememberID(match: Match) {
    for (const id of ["getMatchId", "deleteMatchId"]) {
        (document.getElementById(id) as HTMLInputElement).value = String(match.ID);
    }
}

function describeMatch(match: Match): string {
    const playedAt = match.playedAt ? timestampDate(match.playedAt).toISOString().slice(0, 10) : "?";
    return `#${match.ID} ${match.white} vs ${match.black} ${match.result} `
        + `(${match.moves} moves, ${match.opening ?? "no opening"}, ${playedAt})`;
}

// no id to print, index.Primary("name") makes the name the key
function describePlayer(player: Player): string {
    const joinedAt = player.joinedAt ? timestampDate(player.joinedAt).toISOString().slice(0, 10) : "?";
    return `${player.name} ${player.rating} `
        + `(${player.title ?? "no title"}, joined ${joinedAt})`;
}

function describeStanding(standing: Standing): string {
    return `${standing.ID}. ${standing.player} ${standing.points} pts `
        + `(${standing.played} played, +${standing.wins} =${standing.draws} -${standing.losses})`;
}

// --- Match: both contracts -------------------------------------------------

function createMatch() {
    const white = randomPlayer();
    const black = randomPlayer(white);
    log("Creating match, it is stored in sqlite and returned as proto...");
    const request: StrictMessageInput<CreateMatchRequest> = {
        white: white,
        black: black,
        result: randomResult(),
        opening: randomOpening(),
        moves: randomMoves(),
        playedAt: timestampFromDate(daysAgo(Math.random() * 60)),
    };
    matchClient.create(request)
        .then((response) => {
            log("✓ Match created:", response);
            log(describeMatch(response));
            rememberID(response);
        })
        .catch((error) => {
            log("✗ Error creating match:", error);
        });
}

function createInvalidMatch() {
    const white = randomPlayer();
    log("Creating match with result 2-0, logic.IsKnownResult should reject it...");
    matchClient.create({
        white: white,
        black: randomPlayer(white),
        result: "2-0",
        moves: randomMoves(),
    })
        .then((response) => {
            log("✓ Match created (unexpected):", response);
        })
        .catch((error) => {
            log("✗ Rejected by the generated Validate():", error);
        });
}

function getMatch() {
    const id = numberInput("getMatchId");
    log(`Getting match #${id}...`);
    matchClient.getByID({ ID: id })
        .then((response) => {
            log("✓ Match:", response);
            log(describeMatch(response));
        })
        .catch((error) => {
            log("✗ Error getting match:", error);
        });
}

function deleteMatch() {
    const id = numberInput("deleteMatchId");
    log(`Deleting match #${id}...`);
    matchClient.delete({ ID: id })
        .then(() => {
            log(`✓ Match #${id} deleted, the server wrote an audit row for it`);
        })
        .catch((error) => {
            log("✗ Error deleting match:", error);
        });
}

function listMatches() {
    log("Listing matches from the match table...");
    matchClient.listAll({})
        .then((response) => {
            log(`✓ ${response.matchs.length} matches:`);
            for (const match of response.matchs) {
                log(describeMatch(match));
            }
        })
        .catch((error) => {
            log("✗ Error listing matches:", error);
        });
}

// --- Player: both contracts, proto is read only ----------------------------

function getPlayer() {
    const name = textInput("getPlayerName");
    log(`Getting player ${name}...`);
    playerClient.getByName({ name: name })
        .then((response) => {
            log("✓ Player:", response);
            log(describePlayer(response));
        })
        .catch((error) => {
            log("✗ Error getting player:", error);
        });
}

function listPlayers() {
    log("Listing the roster, seeded by the server on startup...");
    playerClient.listAll({})
        .then((response) => {
            log(`✓ ${response.players.length} players:`);
            for (const player of response.players) {
                log(describePlayer(player));
            }
        })
        .catch((error) => {
            log("✗ Error listing players:", error);
        });
}

// the read only contract leaves create, update and delete off the client
function listPlayerRpcs() {
    const rpcs = Object.keys(PlayerService.method);
    log("Rpcs on PlayerService:", rpcs);
    log("Match has create and delete, Player does not. PROTO().ReadOnly() dropped them.");
    log("CreatePlayer is still a generated query, the server calls it in SeedRoster.");
}

// --- Standing: proto only --------------------------------------------------

function listStandings() {
    log("Listing standings, counted from matches, there is no standings table...");
    standingClient.listAll({})
        .then((response) => {
            log(`✓ ${response.standings.length} players:`);
            for (const standing of response.standings) {
                log(describeStanding(standing));
            }
        })
        .catch((error) => {
            log("✗ Error listing standings:", error);
        });
}

// --- Audit: sqlc only -----------------------------------------------------

function auditCount() {
    log("Audit rows have no rpc, asking the plain /audit-count endpoint instead...");
    fetch("/audit-count")
        .then((response) => response.text())
        .then((count) => {
            log(`✓ ${count} audit rows in sqlite, none of them reachable over grpc`);
        })
        .catch((error) => {
            log("✗ Error reading audit count:", error);
        });
}

// lists what the proto contract actually exposes, Audit is absent
function listServices() {
    const services = Object.keys(schema).filter((name) => name.endsWith("Service"));
    log("Services generated from the proto contract:", services);
    log("Audit declares only entlite.SQLC(), so it has no message and no service");
}

// --- Wiring ---------------------------------------------------------------

document.getElementById("createMatchBtn")!.addEventListener("click", createMatch);
document.getElementById("createInvalidBtn")!.addEventListener("click", createInvalidMatch);
document.getElementById("getMatchBtn")!.addEventListener("click", getMatch);
document.getElementById("deleteMatchBtn")!.addEventListener("click", deleteMatch);
document.getElementById("listMatchesBtn")!.addEventListener("click", listMatches);
document.getElementById("getPlayerBtn")!.addEventListener("click", getPlayer);
document.getElementById("listPlayersBtn")!.addEventListener("click", listPlayers);
document.getElementById("playerRpcsBtn")!.addEventListener("click", listPlayerRpcs);
document.getElementById("listStandingsBtn")!.addEventListener("click", listStandings);
document.getElementById("auditCountBtn")!.addEventListener("click", auditCount);
document.getElementById("listServicesBtn")!.addEventListener("click", listServices);
document.getElementById("clearBtn")!.addEventListener("click", () => {
    document.getElementById("output")!.innerHTML = "";
});

log("Ready. Match both contracts, Player read only proto, Standing only proto, Audit only sqlc.");
