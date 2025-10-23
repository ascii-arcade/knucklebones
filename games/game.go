package games

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/ascii-arcade/knucklebones/database"
	"github.com/ascii-arcade/knucklebones/dice"
	"github.com/ascii-arcade/knucklebones/players"
	"github.com/ascii-arcade/knucklebones/score"
	"github.com/ascii-arcade/knucklebones/utils"
	"github.com/charmbracelet/lipgloss"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Game struct {
	Id         string `bson:"_id"`
	Code       string `bson:"code"`
	InProgress bool   `bson:"in_progress"`
	WinnerID   string `bson:"winner_id,omitempty"`
	Turn       int    `bson:"turn"`

	CreatedAt *time.Time `bson:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at,omitempty"`
	EndedAt   *time.Time `bson:"ended_at,omitempty"`

	// playerOne *players.Player
	// playerTwo *players.Player
	playerOneId string
	playerTwoId string
	players     []struct {
		player *players.Player
		data   *PlayerData
	}

	rolled   bool
	Finished bool

	mu sync.Mutex
}

func (g *Game) Save() error {
	g.UpdatedAt = utils.ToPointer(time.Now())
	collection := database.GetDB().Collection(database.CollectionGames)

	gameJson, err := g.toJson()
	if err != nil {
		return err
	}

	opts := options.Replace().SetUpsert(true)
	_, err = collection.ReplaceOne(database.GetDB().Context(), bson.M{"_id": g.Id}, gameJson, opts)
	return err
}

func (g *Game) toJson() (map[string]any, error) {
	playerIds := make([]string, 0, len(g.players))
	for _, p := range g.players {
		playerIds = append(playerIds, p.player.Id)
	}

	var gameMap map[string]any
	bytes, err := bson.Marshal(g)
	if err != nil {
		slog.Error("error marshalling game to json", "error", err)
		return nil, err
	}
	if err := bson.Unmarshal(bytes, &gameMap); err != nil {
		slog.Error("error unmarshalling game to map", "error", err)
		return nil, err
	}
	gameMap["player_ids"] = playerIds

	return gameMap, nil
}

func (g *Game) refresh() {
	for _, p := range g.players {
		if p.player.IsConnected() && p.data.InGame {
			select {
			case p.player.UpdateChan() <- struct{}{}:
			default:
			}
		}
	}
}

func (g *Game) Rolled() bool {
	r := false
	g.withLock(func() {
		r = g.rolled
	})
	return r
}

func (g *Game) withLock(fn func()) {
	g.mu.Lock()
	defer func() {
		g.refresh()
		g.mu.Unlock()
	}()
	fn()
}

func (g *Game) withErrLock(fn func() error) error {
	g.mu.Lock()
	defer func() {
		g.refresh()
		g.mu.Unlock()
	}()
	return fn()
}

func (g *Game) AddPlayer(player *players.Player) error {
	if g.HasPlayer(player) {
		return nil
	}

	return g.withErrLock(func() error {
		if g.InProgress {
			return ErrGameInProgress
		}

		data := &PlayerData{
			Name:      player.Name,
			Color:     lipgloss.Color(utils.Color()),
			InGame:    true,
			turnOrder: len(g.players),
		}
		data.ResetBoard()
		data.ResetPool()

		player.OnDisconnect(func() {
			if !g.InProgress {
				g.RemovePlayer(player)
			}
		})

		if len(g.players) == 0 {
			data.IsHost = true
			g.players = append(g.players, struct {
				player *players.Player
				data   *PlayerData
			}{
				player: player,
				data:   data,
			})
			return nil
		}

		if len(g.players) == 1 {
			g.players = append(g.players, struct {
				player *players.Player
				data   *PlayerData
			}{
				player: player,
				data:   data,
			})
			return nil
		}

		return ErrGameFull
	})
}

func (g *Game) RemovePlayer(player *players.Player) {
	g.withLock(func() {
		for i, p := range g.players {
			if p.player == player {
				g.players = append(g.players[:i], g.players[i+1:]...)
				break
			}
		}
	})
}

func (g *Game) GetPlayerData(player *players.Player) *PlayerData {
	for _, p := range g.players {
		if p.player == player {
			return p.data
		}
	}
	return nil
}

func (g *Game) GetPlayers() []*players.Player {
	var players []*players.Player
	for _, p := range g.players {
		players = append(players, p.player)
	}
	return players
}

func (g *Game) GetDisconnectedPlayers() []*players.Player {
	var players []*players.Player
	g.withLock(func() {
		for _, p := range g.players {
			if !p.player.IsConnected() {
				players = append(players, p.player)
			}
		}
	})

	if len(players) == 2 {
		g.RemovePlayer(players[0])
		g.RemovePlayer(players[1])
		return nil
	}

	return players
}

func (g *Game) HasPlayer(player *players.Player) bool {
	exists := false
	g.withLock(func() {
		for _, p := range g.players {
			if p.player == player {
				exists = true
				break
			}
		}
	})
	return exists
}

func (g *Game) nextTurn() {
	if g.Turn == 0 {
		g.Turn = 1
	} else {
		g.Turn = 0
	}

	g.rolled = false
}

func (g *Game) GetTurnPlayerData() *PlayerData {
	if g.Turn == 0 {
		return g.players[0].data
	}
	return g.players[1].data
}

func (g *Game) IsTurn(p *players.Player) bool {
	return g.players[g.Turn].player == p
}

func (g *Game) IsPlayerOne(p *players.Player) bool {
	return g.players[0].player == p
}

func (g *Game) GetOpponent(p *players.Player) *players.Player {
	if g.players[0].player == p {
		return g.players[1].player
	}
	return g.players[0].player
}

func (g *Game) GetOpponentData(p *players.Player) *PlayerData {
	if g.players[0].player == p {
		return g.players[1].data
	}
	return g.players[0].data
}

func (g *Game) WinnerData() *PlayerData {
	if !g.Finished {
		return nil
	}

	pOneScore := score.Calculate(g.players[0].data.board)
	pTwoScore := score.Calculate(g.players[1].data.board)
	if pOneScore > pTwoScore {
		return g.players[0].data
	}

	return g.players[1].data
}

func (s *Game) IsPlayerCountOk() error {
	if len(s.players) < 2 {
		return errors.New("not_enough_players")
	}
	return nil
}

func (g *Game) playerOneData() *PlayerData {
	return g.players[0].data
}

func (g *Game) playerTwoData() *PlayerData {
	return g.players[1].data
}

func (g *Game) getTurnPlayer() *players.Player {
	return g.players[g.Turn].player
}

func nextSpot(pool dice.DicePool) int {
	for i, face := range pool {
		if face == 0 {
			return i
		}
	}
	return -1
}

func full(board []dice.DicePool) bool {
	for _, col := range board {
		if nextSpot(col) != -1 {
			return false
		}
	}
	return true
}

func removeSame(column dice.DicePool, face int) {
	for i, n := range column {
		if n == face {
			column[i] = 0 // Remove the die by setting it to 0
		}
	}
}
