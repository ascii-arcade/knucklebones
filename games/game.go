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
	"github.com/charmbracelet/ssh"
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

	playerOne *players.Player
	playerTwo *players.Player

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
	playerOneId := ""
	playerTwoId := ""
	if g.playerOne != nil {
		playerOneId = g.playerOne.Id
	}
	if g.playerTwo != nil {
		playerTwoId = g.playerTwo.Id
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
	gameMap["player_one_id"] = playerOneId
	gameMap["player_two_id"] = playerTwoId

	return gameMap, nil
}

func (g *Game) refresh() {
	players := []*players.Player{g.playerOne, g.playerTwo}
	for _, p := range players {
		if p != nil && p.UpdateChan != nil {
			select {
			case p.UpdateChan <- struct{}{}:
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
	return g.withErrLock(func() error {
		if _, ok := g.getPlayer(player.Sess); ok {
			return nil
		}

		if g.InProgress {
			return ErrGameInProgress
		}

		player.OnDisconnect(func() {
			if !g.InProgress {
				g.RemovePlayer(player)
			}
		})

		if g.playerOne == nil {
			player.MakeHost()
			g.playerOne = player
			return nil
		}

		if g.playerTwo == nil {
			g.playerTwo = player
			return nil
		}

		return nil
	})
}

func (g *Game) RemovePlayer(player *players.Player) {
	g.withLock(func() {
		if player, exists := g.getPlayer(player.Sess); exists {
			close(player.UpdateChan)
			if g.playerOne == player {
				g.playerOne = nil
			} else if g.playerTwo == player {
				g.playerTwo = nil
			}
		}
	})
}

func (g *Game) getPlayer(sess ssh.Session) (*players.Player, bool) {
	if g.playerOne != nil && g.playerOne.Sess.User() == sess.User() {
		return g.playerOne, true
	} else if g.playerTwo != nil && g.playerTwo.Sess.User() == sess.User() {
		return g.playerTwo, true
	}
	return nil, false
}

func (g *Game) GetPlayers() []*players.Player {
	var players []*players.Player
	if g.playerOne != nil {
		players = append(players, g.playerOne)
	}
	if g.playerTwo != nil {
		players = append(players, g.playerTwo)
	}
	return players
}

func (g *Game) GetDisconnectedPlayers() []*players.Player {
	var players []*players.Player
	g.withLock(func() {
		if !g.playerOne.Connected {
			players = append(players, g.playerOne)
		}
		if !g.playerTwo.Connected {
			players = append(players, g.playerTwo)
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
	_, exists := g.getPlayer(player.Sess)
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

func (g *Game) GetTurnPlayer() *players.Player {
	if g.Turn == 0 {
		return g.playerOne
	}
	return g.playerTwo
}

func (g *Game) IsTurn(p *players.Player) bool {
	return g.GetTurnPlayer().Name == p.Name
}

func (g *Game) IsPlayerOne(p *players.Player) bool {
	return g.playerOne.Name == p.Name
}

func (g *Game) GetOpponent(p *players.Player) *players.Player {
	if g.playerOne == p {
		return g.playerTwo
	}
	return g.playerOne
}

func (g *Game) Winner() *players.Player {
	if !g.Finished {
		return nil
	}

	pOneScore := score.Calculate(g.playerOne.Board)
	pTwoScore := score.Calculate(g.playerTwo.Board)
	if pOneScore > pTwoScore {
		return g.playerOne
	}

	return g.playerTwo
}

func (s *Game) IsPlayerCountOk() error {
	if s.playerTwo == nil {
		return errors.New("not_enough_players")
	}
	return nil
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
