package games

import (
	"errors"

	"github.com/ascii-arcade/knucklebones/players"
	"github.com/ascii-arcade/knucklebones/utils"
)

var (
	ErrGameInProgress = errors.New("game_already_in_progress")
	ErrGameNotFound   = errors.New("game_not_found")
	ErrGameFull       = errors.New("game_full")
	ErrNotYourTurn    = errors.New("not_your_turn")
	ErrDiceNotRolled  = errors.New("dice_not_rolled")
	ErrColumnFull     = errors.New("column_full")
)

var games = make(map[string]*Game)

func New() *Game {
	game := &Game{
		Code: utils.Code(),
		players: []struct {
			player *players.Player
			data   *PlayerData
		}{},
	}
	games[game.Code] = game

	return game
}

func GetOpenGame(code string) (*Game, error) {
	game, exists := games[code]
	if !exists {
		return nil, ErrGameNotFound
	}
	if game.InProgress {
		return game, ErrGameInProgress
	}

	return game, nil
}

func GetAll() map[string]*Game {
	return games
}
