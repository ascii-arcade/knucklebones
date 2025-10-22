package games

import (
	"log/slog"
	"time"

	"github.com/ascii-arcade/knucklebones/database"
	"github.com/ascii-arcade/knucklebones/dice"
	"github.com/ascii-arcade/knucklebones/players"
)

type ActionType string

const (
	ActionTypeStart ActionType = "start"
	ActionTypeRoll  ActionType = "roll"
	ActionTypePlace ActionType = "place"
)

type Action struct {
	GameID     string          `bson:"game_id"`
	PlayerID   string          `bson:"player_id"`
	Action     ActionType      `bson:"action"`
	DicePool   dice.DicePool   `bson:"dice_pool,omitempty"`
	DiceLocked []dice.DicePool `bson:"dice_locked,omitempty"`
	Turn       int             `bson:"turn"`
	Timestamp  time.Time       `bson:"time"`
}

func (a *Action) Save() error {
	collection := database.GetDB().Collection(database.CollectionActions)
	_, err := collection.InsertOne(database.GetDB().Context(), a)
	return err
}

func (g *Game) saveAction(actionType ActionType) {
	action := Action{
		GameID:    g.Id,
		PlayerID:  g.GetTurnPlayer().Id,
		Action:    actionType,
		Turn:      g.Turn,
		Timestamp: time.Now().UTC(),
	}
	if err := action.Save(); err != nil {
		slog.Error("error saving action", "error", err)
	}

	if err := g.Save(); err != nil {
		slog.Error("error saving game after action", "error", err)
	}
}

func (g *Game) Begin() error {
	return g.withErrLock(func() error {
		if error := g.IsPlayerCountOk(); error != nil {
			return error
		}

		g.InProgress = true
		g.saveAction(ActionTypeStart)
		return nil
	})
}

func (g *Game) Reset() {
	g.withLock(func() {
		g.playerOne.Board = make([]dice.DicePool, 3)
		g.playerTwo.Board = make([]dice.DicePool, 3)
		for i := range g.playerOne.Board {
			g.playerOne.Board[i] = make(dice.DicePool, 3)
			g.playerTwo.Board[i] = make(dice.DicePool, 3)
		}

		g.playerOne.Pool = make(dice.DicePool, 1)
		g.playerTwo.Pool = make(dice.DicePool, 1)

		g.rolled = false
		g.Finished = false
		g.Turn = 0
	})
}

func (g *Game) RollDice(rolling bool) {
	g.withLock(func() {
		switch g.Turn {
		case 0:
			g.playerOne.Pool.Roll()
		case 1:
			g.playerTwo.Pool.Roll()
		}

		if !rolling {
			g.rolled = true
		}

		g.saveAction(ActionTypeRoll)
	})
}

func (g *Game) PlaceDie(p *players.Player, column int) error {
	return g.withErrLock(func() error {
		if !g.IsTurn(p) {
			return ErrNotYourTurn
		}

		if !g.rolled {
			return ErrDiceNotRolled
		}

		spot := nextSpot(p.Board[column])
		if spot == -1 {
			return ErrColumnFull
		}

		p.Board[column][spot] = p.Pool[0]
		p.Pool = make(dice.DicePool, 1)

		removeSame(g.GetOpponent(p).Board[column], p.Board[column][spot])

		if full(p.Board) {
			g.Finished = true
			return nil
		}

		g.nextTurn()
		g.saveAction(ActionTypePlace)
		return nil
	})
}
