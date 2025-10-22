package games

import (
	"errors"

	"github.com/ascii-arcade/knucklebones/generaterandom"
)

var (
	ErrGameInProgress = errors.New("game_already_in_progress")
	ErrGameNotFound   = errors.New("game_not_found")
	ErrNotYourTurn    = errors.New("not_your_turn")
	ErrDiceNotRolled  = errors.New("dice_not_rolled")
	ErrColumnFull     = errors.New("column_full")
)

var games = make(map[string]*Game)

func New() *Game {
	game := &Game{
		Code: generaterandom.Code(),
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
