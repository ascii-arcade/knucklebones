package games

import (
	"github.com/ascii-arcade/knucklebones/dice"
	"github.com/charmbracelet/lipgloss"
)

type PlayerData struct {
	Name           string
	Score          int
	Color          lipgloss.Color
	PlayedLastTurn bool
	InGame         bool

	IsHost bool

	turnOrder int
	board     []dice.DicePool
	pool      dice.DicePool
}

func (p *PlayerData) ResetBoard() {
	p.board = make([]dice.DicePool, 3)
	for i := range p.board {
		p.board[i] = make(dice.DicePool, 3)
	}
	p.pool = make(dice.DicePool, 1)
}

func (p *PlayerData) Board() []dice.DicePool {
	return p.board
}

func (p *PlayerData) ResetPool() {
	p.pool = make(dice.DicePool, 1)
}

func (p *PlayerData) Pool() *dice.DicePool {
	return &p.pool
}

func (p *PlayerData) StyledPlayerName(style lipgloss.Style) string {
	return style.Foreground(p.Color).Render(p.Name)
}
